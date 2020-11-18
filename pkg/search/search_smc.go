package search

import (
	"encoding/xml"
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"github.com/tstromberg/campwiz/pkg/cache"
	"github.com/tstromberg/campwiz/pkg/geo"
	"k8s.io/klog"
)

var (
	smcSiteCodes = []string{"coyote-point"}
	smcRoot      = "secure" + ".itinio" + ".com" + "/sanmateo"
	metaSmcRoot  = "u/us/ca/san_mateo/"
	smcCenterLon = 37.4250399
	smcCenterLat = -122.4130398
)

// searchSMC searches San Mateo County Parks for a single date
func searchSMC(q Query, date time.Time, cs cache.Store) ([]Result, error) {
	dist := geo.MilesApart(q.Lat, q.Lon, smcCenterLat, smcCenterLon)
	klog.Infof("searchSMC, distance to center: %.1f miles", dist)
	if dist > float64(q.MaxDistance) {
		klog.Warningf("skipping smc search -- too far away")
		return nil, nil
	}

	klog.Infof("searchSMC: %+v", q)
	var results []Result

	for _, code := range smcSiteCodes {
		result, cached, err := checkSMCSite(code, q, date, cs)
		if err != nil {
			return nil, fmt.Errorf("smc site: %w", err)
		}

		if result != nil {
			results = append(results, *result)
		}
		if !cached {
			klog.V(1).Infof("Previous request was uncached, sleeping ...")
			time.Sleep(uncachedDelay)
		}
	}
	klog.Infof("returning %d results", len(results))
	return results, nil
}

func smcSiteURL(code string) string {
	return "https://" + smcRoot + "/" + code
}

// endDate returns a calculated end date
func endDate(start time.Time, stayLength int) time.Time {
	return start.Add(time.Duration(stayLength) * 24 * time.Hour)
}

func smcSiteRequest(code string, q Query, date time.Time) cache.Request {
	// https://secure.itinio.com/sanmateo/campsites/feed.html?startDate=2021-01-01&endDate=2021-01-02&code=0.9170750459654033

	v := url.Values{
		"startDate": {date.Format("2006-01-02")},
		"endDate":   {endDate(date, q.StayLength).Format("2006-01-02")},
		"code":      {fmt.Sprintf("%0.16f", rand.Float64())}, // Weird, but this is what SMC expects!
	}

	r := cache.Request{
		Method:   "GET",
		URL:      smcRoot + "/feed.html",
		Referrer: smcSiteURL(code),
		Form:     v,
		MaxAge:   searchPageExpiry,
	}
	return r
}

func checkSMCSite(code string, q Query, date time.Time, cs cache.Store) (*Result, bool, error) {
	sr, err := cache.Fetch(smcSiteRequest(code, q, date), cs)
	if err != nil {
		return nil, false, fmt.Errorf("fetch start: %w", err)
	}

	cookies := sr.Cookies
	klog.Infof("start page cached: %v cookies: %+v", sr.Cached, sr.Cookies)

	req := smcSiteRequest(code, q, date)
	req.Cookies = cookies

	resp, err := cache.Fetch(req, cs)
	if err != nil {
		return nil, false, fmt.Errorf("fetch: %w", err)
	}

	result, err := parseSMCSearchPage(code, resp.Body, date, q)
	return result, resp.Cached, err
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

func parseSMCSearchPage(code string, bs []byte, date time.Time, q Query) (*Result, error) {
	var sites Sites

	err := xml.Unmarshal(bs, &sites)
	if err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	klog.V(2).Infof("unmarshalled data: %+v", sites)

	for _, s := range sites.Sites {
		if s.Available != 1 {
			continue
		}
		a := Availability{
			SiteType: "campsite",
			Date:     date,
			URL:      smcSiteURL(code),
		}

		r := &Result{
			ID:           "parks.smcgov.org/" + code,
			Name:         "nothing yet",
			Distance:     geo.MilesApart(q.Lat, q.Lon, smcCenterLat, smcCenterLon),
			Availability: []Availability{a},
		}

		klog.Infof("%s is available: %+v", r.Name, r)
		return r, nil
	}

	return nil, nil
}
