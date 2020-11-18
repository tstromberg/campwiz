package provider

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"time"

	"github.com/peterbourgon/diskv"
	"github.com/tstromberg/campwiz/pkg/cache"
	"github.com/tstromberg/campwiz/pkg/geo"
	"k8s.io/klog"
)

var (
	smcSiteCodes = map[string]string{
		"coyote-point": "Coyote Point Recreation Area",
	}
	smcRoot      = "secure" + ".itinio" + ".com" + "/sanmateo"
	smcCenterLon = 37.4250399
	smcCenterLat = -122.4130398
)

func smcSiteURL(code string) string {
	return "https://" + smcRoot + "/" + code
}

func smcSiteRequest(code string, q Query, date time.Time) cache.Request {
	// https://secure.itinio.com/sanmateo/campsites/feed.html?startDate=2021-01-01&endDate=2021-01-02&code=0.9170750459654033

	v := url.Values{
		"startDate": {date.Format("2006-01-02")},
		"endDate":   {date.Add(24 * time.Hour * q.StayLength).Format("2006-01-02")},
		"code":      math.Random(1),
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

func checkSMCSite(code string, q Query, date time.Time, dv *diskv.Diskv) (Result, bool, error) {
	sr, err := cache.Fetch(smcSiteRequest(code, q, date), dv)
	if err != nil {
		return nil, false, fmt.Errorf("fetch start: %w", err)
	}

	cookies := sr.Cookies
	klog.Infof("start page cached: %v cookies: %+v", sr.Cached, sr.Cookies)

	req := smcSiteRequest(q, date, i)
	req.Cookies = cookies

	resp, err := cache.Fetch(req, dv)
	if err != nil {
		return nil, false, fmt.Errorf("fetch: %w", err)
	}

	result := parseSMCSearchPage(code, resp.Body, date, q)
	return result, resp.Cached, nil
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

func parseSMCSearchPage(code string, bs []byte, date time.Time, q Query) (Result, error) {
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

		rr := Result{
			ID:           "smcgov.org/" + code,
			Name:         smcSiteCodes[code],
			Distance:     geo.MilesApart(q.Lat, q.Lon, smcCenterLat, smcCenterLon),
			Availability: []Availability{a},
		}

		klog.Infof("%s is available: %+v", r.Name, rr)
	}
}

// searchSMC runs a search for a single date
func searchSMC(q Query, date time.Time, dv *diskv.Diskv) ([]Result, error) {
	dist := geo.MilesApart(q.Lat, q.Lon, smcCenterLat, smcCenterLon)
	klog.Infof("searchSMC, distance to center: %s miles", dist)
	if dist > 75 {
		klog.Warningf("skipping smc search -- too far away")
		return nil, nil
	}

	klog.Infof("searchSMC: %+v", q)
	var results []Result

	for _, code := range smcSiteCodes {
		result, cached, err := checkSMCSite(code, q, date, dv)
		if err != nil {
			return nil, fmt.Errorf("smc site: %w", err)
		}

		if result != nil {
			results = append(results, result)
		}
		if !resp.Cached {
			klog.V(1).Infof("Previous request was uncached, sleeping ...")
			time.Sleep(uncachedDelay)
		}
	}
	klog.Infof("returning %d results", len(results))
	return results, nil
}
