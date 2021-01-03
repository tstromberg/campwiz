package backend

import (
	"bytes"
	"fmt"
	"net/http/cookiejar"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/tstromberg/campwiz/pkg/cache"
	"github.com/tstromberg/campwiz/pkg/campwiz"
	"github.com/tstromberg/campwiz/pkg/geo"
	"github.com/tstromberg/campwiz/pkg/mangle"
	"k8s.io/klog/v2"
)

var (
	// sccCenterLat is the center of Santa Clara County, used for approximate location filtering
	sccCenterLat = 37.1908873
	// sccCenterLon is the center of Santa Clara County, used for approximate location filtering
	sccCenterLon = -122.4130398
)

// SantaClaraCounty handles SantaClaraCounty queries
type SantaClaraCounty struct {
	store cache.Store
	jar   *cookiejar.Jar
}

// Name is a human readable name
func (b *SantaClaraCounty) Name() string {
	return "Santa Clara County Parks"
}

// List lists available sites
func (b *SantaClaraCounty) List(q campwiz.Query) ([]campwiz.Result, error) {
	klog.Infof("SantaClaraCounty.List: %+v", q)
	_, err := cache.Fetch(b.startPage(), b.store)
	if err != nil {
		return nil, fmt.Errorf("fetch start: %w", err)
	}

	var res []campwiz.Result
	for _, d := range q.Dates {
		rs, err := b.avail(q, d)
		if err != nil {
			return res, fmt.Errorf("avail: %w", err)
		}
		res = append(res, rs...)
	}

	return mergeDates(res), nil
}

// url is the root URL to use for requests
func (b *SantaClaraCounty) url(s string) string {
	return "https://" + "gooutsideandplay" + ".org" + s
}

// req generates a search request
func (b *SantaClaraCounty) req(q campwiz.Query, arrival time.Time) cache.Request {
	firstBook := time.Now().Add(24 * time.Hour)
	lastBook := time.Now().Add(6 * 30 * 24 * time.Hour)
	v := url.Values{
		"actiontype":                {"camping"},
		"park_idno":                 {"0"},
		"CalendarCurrentDate":       {time.Now().Format("01/02/2006")},
		"CalendarFirstBookableDate": {firstBook.Format("01/02/2006")},
		"CalendarLastBookableDate":  {lastBook.Format("01/02/2006")},
		"use_type":                  {""},
		"res_length":                {strconv.Itoa(q.StayLength)},
		"arrive_date":               {arrival.Format("01/02/2006")},
		"c_park_idno":               {"0"},
		"d_park_idno":               {"0"},
		"b_park_idno":               {"1"},
		"center_idno":               {"0"},
		"facility_use_type_idno":    {"0"},
	}
	klog.Infof("campwiz.Query: %+v", q)
	klog.Infof("values: %+v", v)

	r := cache.Request{
		Method:   "GET",
		URL:      b.url("/index.asp"),
		Referrer: b.url("/"),
		Form:     v,
		Jar:      b.jar,
	}
	return r
}

// searchReq generates an initial page request
func (b *SantaClaraCounty) startPage() cache.Request {
	return cache.Request{URL: b.url("/index.asp"), Referrer: b.url("/"), Jar: b.jar}
}

// parse parses the search response
func (b *SantaClaraCounty) parse(bs []byte, date time.Time, q campwiz.Query) ([]campwiz.Result, error) {
	// name to result
	sites := map[string]*campwiz.Result{}
	// name+kind to avail
	avail := map[string]map[string]*campwiz.Availability{}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(bs))
	if err != nil {
		return nil, fmt.Errorf("new doc: %w", err)
	}

	listing := doc.Find("#list_camping")

	// Find the review items
	listing.Find("tr").Each(func(i int, s *goquery.Selection) {
		h, err := goquery.OuterHtml(s)
		if err != nil {
			klog.Errorf("no html: %v", err)
			return
		}

		name := s.Find(".body_gray").Text()
		if name == "" {
			klog.Warningf("no name within: %s", h)
			return
		}

		sid := s.Find(".heavy_blue").Text()
		if name == "" {
			klog.Warningf("no sid within: %s", h)
			return
		}

		sid = strings.TrimSpace(sid)
		sType := s.Find(".body_blue").Text()

		klog.Infof("name: %s type: %s sid: %s", name, sType, sid)

		_, ok := avail[name]
		if !ok {
			avail[name] = map[string]*campwiz.Availability{}
		}

		sKind := mangle.SiteKind(name, sType, sid)

		// Group availability by type + kind (may differ based on site id)
		availKey := fmt.Sprintf("%s=%s", sType, sKind)
		a, ok := avail[name][availKey]
		if ok {
			a.SpotCount++
			return
		}

		// TODO: Support multiple types + populate count
		avail[name][availKey] = &campwiz.Availability{
			SiteKind:  sKind,
			SiteDesc:  sType,
			Name:      name,
			Date:      date,
			SpotCount: 1,
			URL:       b.url(s.Find(".FilterElement a").AttrOr("href", "")),
		}

		sites[name] = &campwiz.Result{
			ResURL:   b.url("/"),
			ResID:    strings.ToLower(strings.Replace(name, " ", "_", -1)),
			Name:     name,
			Distance: geo.MilesApart(q.Lat, q.Lon, sccCenterLat, sccCenterLon),
		}

	})

	// combine everything
	results := []campwiz.Result{}
	for _, r := range sites {
		for _, a := range avail[r.Name] {
			r.Availability = append(r.Availability, *a)
		}

		sort.Slice(r.Availability, func(i, j int) bool {
			return string(r.Availability[i].SiteKind)+r.Availability[i].SiteDesc < string(r.Availability[j].SiteKind)+r.Availability[j].SiteDesc
		})
		results = append(results, *r)
	}

	sort.Slice(results, func(i, j int) bool { return results[i].Name < results[j].Name })

	return results, nil
}

// avail lists sites available on a single date
func (b *SantaClaraCounty) avail(q campwiz.Query, d time.Time) ([]campwiz.Result, error) {
	dist := geo.MilesApart(q.Lat, q.Lon, sccCenterLat, sccCenterLon)
	klog.Infof("searchSCC, distance to center from %f / %f is %.1f miles", q.Lat, q.Lon, dist)
	if dist > float64(q.MaxDistance) {
		klog.Warningf("skipping scc search -- further than %d miles", q.MaxDistance)
		return nil, nil
	}

	_, err := cache.Fetch(b.startPage(), b.store)
	if err != nil {
		return nil, fmt.Errorf("fetch start: %w", err)
	}

	req := b.req(q, d)
	resp, err := cache.Fetch(req, b.store)
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}

	prs, err := b.parse(resp.Body, d, q)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	return prs, err
}
