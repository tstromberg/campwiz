package search

import (
	"bytes"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/tstromberg/campwiz/pkg/cache"
	"github.com/tstromberg/campwiz/pkg/geo"
	"k8s.io/klog"
)

var (
	sccRoot      = "https://" + "gooutsideandplay" + ".org"
	metaSccRoot  = "/u/us/ca/santa_clara/"
	sccCenterLat = 37.1908873
	sccCenterLon = -122.4130398
)

// searchSCC searches San Mateo County Parks for a single date
func searchSCC(q Query, date time.Time, cs cache.Store) ([]Result, error) {
	dist := geo.MilesApart(q.Lat, q.Lon, sccCenterLat, sccCenterLon)
	klog.Infof("searchSCC, distance to center from %f / %f is %.1f miles", q.Lat, q.Lon, dist)
	if dist > float64(q.MaxDistance) {
		klog.Warningf("skipping scc search -- further than %d miles", q.MaxDistance)
		return nil, nil
	}

	sr, err := cache.Fetch(sccStartPage(), cs)
	if err != nil {
		return nil, fmt.Errorf("fetch start: %w", err)
	}

	cookies := sr.Cookies
	klog.Infof("start page cached: %v cookies: %+v", sr.Cached, sr.Cookies)

	req := sccSiteRequest(q, date)
	req.Cookies = cookies

	resp, err := cache.Fetch(req, cs)
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}

	results, err := parseSCCSearchPage(resp.Body, date, q)
	return results, err
}

// sccStartPage returns a request for the search page
func sccStartPage() cache.Request {
	return cache.Request{
		Method: "GET",
		URL:    sccRoot + "/index.asp",
		MaxAge: searchPageExpiry,
	}
}

func sccSiteRequest(q Query, date time.Time) cache.Request {
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
		"arrive_date":               {date.Format("01/02/2006")},
		"c_park_idno":               {"0"},
		"d_park_idno":               {"0"},
		"b_park_idno":               {"1"},
		"center_idno":               {"0"},
		"facility_use_type_idno":    {"0"},
	}
	klog.Infof("query: %+v", q)
	klog.Infof("values: %+v", v)

	r := cache.Request{
		Method:   "GET",
		URL:      sccRoot + "/index.asp",
		Referrer: sccRoot,
		Form:     v,
		MaxAge:   searchPageExpiry,
	}
	return r
}

func parseSCCSearchPage(bs []byte, date time.Time, q Query) ([]Result, error) {
	var results []Result

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
		a := Availability{
			SiteType: stype,
			Date:     date,
			URL:      sccRoot + s.Find(".FilterElement a").AttrOr("href", ""),
		}

		r := Result{
			ID:           metaSccRoot + strings.ToLower(strings.Replace(name, " ", "_", -1)),
			Name:         name,
			Distance:     geo.MilesApart(q.Lat, q.Lon, sccCenterLat, sccCenterLon),
			Availability: []Availability{a},
		}

		results = append(results, r)
	})

	return results, nil
}
