package backend

import (
	"bytes"
	"fmt"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/tstromberg/campwiz/pkg/cache"
	"github.com/tstromberg/campwiz/pkg/campwiz"
	"github.com/tstromberg/campwiz/pkg/geo"
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
	return "https://" + "gooutsideandplay" + ".org"
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
	}
	return r
}

// searchReq generates an initial page request
func (b *SantaClaraCounty) startPage() cache.Request {
	return cache.Request{URL: b.url("/index.asp"), Referrer: b.url("/"), Jar: b.jar}
}

// parse parses the search response
func (b *SantaClaraCounty) parse(bs []byte, date time.Time, q campwiz.Query) ([]campwiz.Result, error) {
	var results []campwiz.Result

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(bs))
	if err != nil {
		return results, fmt.Errorf("new doc: %w", err)
	}

	listing := doc.Find("#list_camping")

	seen := map[string]bool{}

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

		stype := s.Find(".body_blue").Text()

		klog.Infof("name: %s type: %s", name, stype)
		if seen[name+stype] {
			return
		}

		// TODO: Support multiple types + populate count
		seen[name+stype] = true
		a := campwiz.Availability{
			SiteType: stype,
			Date:     date,
			URL:      b.url(s.Find(".FilterElement a").AttrOr("href", "")),
		}

		r := campwiz.Result{
			ResURL:       b.url("/"),
			ResID:        strings.ToLower(strings.Replace(name, " ", "_", -1)),
			Name:         name,
			Distance:     geo.MilesApart(q.Lat, q.Lon, sccCenterLat, sccCenterLon),
			Availability: []campwiz.Availability{a},
		}

		results = append(results, r)
	})

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
