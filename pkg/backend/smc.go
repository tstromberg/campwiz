package backend

import (
	"encoding/xml"
	"fmt"
	"math/rand"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/tstromberg/campwiz/pkg/cache"
	"github.com/tstromberg/campwiz/pkg/campwiz"
	"github.com/tstromberg/campwiz/pkg/geo"
	"k8s.io/klog/v2"
)

var (
	// San Mateo only has two camping sites available for reservation at the moment
	smcSiteIDs   = []string{"coyote-point", "huddart-park"}
	smcCenterLat = 37.4250399
	smcCenterLon = -122.4130398
)

// SanMateoCounty handles Santa Mateo County Parks queries
type SanMateoCounty struct {
	store cache.Store
	jar   *cookiejar.Jar
}

// Name is a human readable name
func (b *SanMateoCounty) Name() string {
	return "San Mateo County"
}

// List lists available sites
func (b *SanMateoCounty) List(q campwiz.Query) ([]campwiz.Result, error) {

	var res []campwiz.Result
	for _, siteID := range smcSiteIDs {
		_, err := cache.Fetch(b.startPage(siteID), b.store)
		if err != nil {
			return nil, fmt.Errorf("fetch start: %w", err)
		}

		for _, d := range q.Dates {
			rs, err := b.avail(q, d, siteID)
			if err != nil {
				return res, fmt.Errorf("avail: %w", err)
			}
			res = append(res, rs...)
		}
	}

	return mergeDates(res), nil
}

// url is the root URL to use for requests
func (b *SanMateoCounty) url(s string) string {
	return "https://" + "secure" + ".itinio" + ".com" + "/sanmateo" + s
}

// startPage generates an initial page request
func (b *SanMateoCounty) startPage(siteID string) cache.Request {
	return cache.Request{URL: b.url("/" + siteID), Referrer: b.url("/"), Jar: b.jar}
}

// req generates a search request
func (b *SanMateoCounty) req(q campwiz.Query, arrival time.Time, siteID string) cache.Request {
	v := url.Values{
		"startDate": {arrival.Format("2006-01-02")},
		"endDate":   {endDate(arrival, q.StayLength).Format("2006-01-02")},
		"code":      {fmt.Sprintf("%0.16f", rand.Float64())}, // Weird, but this is what SMC expects!
	}

	r := cache.Request{
		Method:   "GET",
		URL:      b.url("/campsites/feed.html"),
		Referrer: b.url("/" + siteID),
		Form:     v,
		MaxAge:   searchPageExpiry,
		Jar:      b.jar,
	}
	return r
}

type Sites struct {
	XMLName xml.Name `xml:"sites"`
	Sites   []Site   `xml:"site"`
}

type Site struct {
	XMLName   xml.Name `xml:"site"`
	SiteID    string   `xml:"siteId,attr"`
	Available int      `xml:"avail,attr"`
}

func siteIDToTitle(s string) string {
	return strings.Title(strings.Replace(s, "-", " ", -1))
}

// parse parses the search response
func (b *SanMateoCounty) parse(bs []byte, date time.Time, q campwiz.Query, siteID string) ([]campwiz.Result, error) {
	var sites Sites
	var results []campwiz.Result

	err := xml.Unmarshal(bs, &sites)
	if err != nil {
		return nil, fmt.Errorf("unmarshal: %w\ncontent: %s", err, bs)
	}

	klog.V(2).Infof("unmarshalled data: %+v", sites)

	for _, s := range sites.Sites {
		if s.Available != 1 {
			continue
		}
		a := campwiz.Availability{
			SiteType: "campsite",
			Date:     date,
			URL:      b.url("/" + siteID),
		}

		r := campwiz.Result{
			ResID:        siteID,
			ResURL:       b.url("/"),
			Name:         siteIDToTitle(siteID),
			Distance:     geo.MilesApart(q.Lat, q.Lon, smcCenterLat, smcCenterLon),
			Availability: []campwiz.Availability{a},
		}

		klog.Infof("%s is available: %+v", r.Name, r)
		results = append(results, r)
	}

	return results, nil
}

// avail lists sites available on a single date / location
func (b *SanMateoCounty) avail(q campwiz.Query, d time.Time, siteID string) ([]campwiz.Result, error) {
	dist := geo.MilesApart(q.Lat, q.Lon, smcCenterLat, smcCenterLon)
	klog.Infof("searchSMC, distance to center from %f / %f is %.1f miles", q.Lat, q.Lon, dist)
	if dist > float64(q.MaxDistance) {
		klog.Warningf("skipping smc search -- further than %d miles", q.MaxDistance)
		return nil, nil
	}

	req := b.req(q, d, siteID)
	resp, err := cache.Fetch(req, b.store)
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}

	prs, err := b.parse(resp.Body, d, q, siteID)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	return prs, err
}
