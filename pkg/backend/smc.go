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
	smcSiteCodes = []string{"coyote-point", "huddart-park"}
	smcRoot      = "https://" + "secure" + ".itinio" + ".com" + "/sanmateo"
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
	return nil, nil
}

// searchSMC searches San Mateo County Parks for a single date
func searchSMC(q campwiz.Query, date time.Time, cs cache.Store) ([]campwiz.Result, error) {
	dist := geo.MilesApart(q.Lat, q.Lon, smcCenterLat, smcCenterLon)
	klog.Infof("searchSMC, distance to center from %f / %f is %.1f miles", q.Lat, q.Lon, dist)
	if dist > float64(q.MaxDistance) {
		klog.Warningf("skipping smc search -- further than %d miles", q.MaxDistance)
		return nil, nil
	}

	klog.Infof("searchSMC: %+v", q)
	var results []campwiz.Result

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
	return smcRoot + "/" + code
}

// endDate returns a calculated end date
func endDate(start time.Time, stayLength int) time.Time {
	return start.Add(time.Duration(stayLength) * 24 * time.Hour)
}

// smcStartPage returns a request for the search page
func smcStartPage(code string) cache.Request {
	return cache.Request{
		Method: "GET",
		URL:    smcSiteURL(code),
		MaxAge: searchPageExpiry,
	}
}

func smcSiteRequest(code string, q campwiz.Query, date time.Time) cache.Request {
	// https://secure.itinio.com/sanmateo/campsites/feed.html?startDate=2021-01-01&endDate=2021-01-02&code=0.9170750459654033

	v := url.Values{
		"startDate": {date.Format("2006-01-02")},
		"endDate":   {endDate(date, q.StayLength).Format("2006-01-02")},
		"code":      {fmt.Sprintf("%0.16f", rand.Float64())}, // Weird, but this is what SMC expects!
	}

	r := cache.Request{
		Method:   "GET",
		URL:      smcRoot + "/campsites/feed.html",
		Referrer: smcSiteURL(code),
		Form:     v,
		MaxAge:   searchPageExpiry,
	}
	return r
}

func checkSMCSite(code string, q campwiz.Query, date time.Time, cs cache.Store) (*campwiz.Result, bool, error) {
	sr, err := cache.Fetch(smcStartPage(code), cs)
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

func codeToTitle(s string) string {
	return strings.Title(strings.Replace(s, "-", " ", -1))
}

func parseSMCSearchPage(code string, bs []byte, date time.Time, q campwiz.Query) (*campwiz.Result, error) {
	var sites Sites

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
			URL:      smcSiteURL(code),
		}

		r := &campwiz.Result{

			ResID: code,
			ResURL: smcSiteURL("/"),
			Name:         codeToTitle(code),
			Distance:     geo.MilesApart(q.Lat, q.Lon, smcCenterLat, smcCenterLon),
			Availability: []campwiz.Availability{a},
		}

		klog.Infof("%s is available: %+v", r.Name, r)
		return r, nil
	}

	return nil, nil
}
