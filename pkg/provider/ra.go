package provider

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/tstromberg/campwiz/pkg/cache"
	"k8s.io/klog/v2"
)

var (
	// raURL is the search URL to request reservation information from.
	raURL = "https://" + "www." + "reserve" + "america.com"

	// searchPageExpiry is how long search pages can be cached for.
	searchPageExpiry = time.Duration(6*3600) * time.Second

	// amount of time to sleep between uncached fetches
	uncachedDelay = time.Millisecond * 300

	// maximum number of pages to fetch
	maxPages = 15
)

// pageRequest creates the request object for a search.
func pageRequest(c Query, arrival time.Time, num int) cache.Request {
	// https://www.reserveamerica.com/jaxrs-json/search?rcp=0&stype=nearby&lng=-122.443&lat=37.7562&arv=2021-02-05&lsy=2&pa99999=2003&pa12=4&pa24=true&interest=camping&rcs=20

	v := url.Values{
		"rcp":     {strconv.Itoa(num)},            // page number
		"stype":   {"nearby"},                     // search type
		"lng":     {fmt.Sprintf("%3.3f", c.Lon)},  // Longitude
		"lat":     {fmt.Sprintf("%3.3f", c.Lat)},  // Latitude
		"arv":     {arrival.Format("2006-01-02")}, // arrival date,
		"lsy":     {strconv.Itoa(c.StayLength)},   // length of stay
		"pa99999": {"2003"},                       // looking for (tent). See https://developer.active.com/docs/read/Campground_Search_API
		// "pa12": # of people
		// "pa24": waterfront
		"rcs":      {"100"}, // 100 results
		"interest": {"camping"},
	}

	r := cache.Request{
		Method:   "GET",
		URL:      raURL + "/jaxrs-json/search",
		Referrer: raURL,
		Form:     v,
		MaxAge:   searchPageExpiry,
	}

	for v, k := range v {
		klog.Infof("Form value %s = %q", v, k)
	}
	return r
}

// startPage returns a request for the search page
func startPage() cache.Request {
	return cache.Request{
		Method:   "GET",
		URL:      raURL + "/explore/search-results",
		Referrer: raURL,
		MaxAge:   searchPageExpiry,
	}
}

type jaxControl struct {
	CurrentPage int
	PageSize    int
}

type jaxRecord struct {
	NamingID  string
	Name      string
	Proximity float32
	Details   jaxDetails
}

type jaxAvailability struct {
	Available      bool
	ReservableType string
}

type jaxDetails struct {
	ID           int32
	ContrCode    string
	raURL        string
	Availability jaxAvailability
}

type jaxResponse struct {
	TotalRecords int
	TotalPages   int
	StartIndex   int
	EndIndex     int
	Control      jaxControl
	Records      []jaxRecord
}

// searchRA runs a search for a single date
func searchRA(crit Query, date time.Time) ([]Result, error) {
	klog.Infof("searchRA: %+v", crit)

	// grab the current cookies
	sr, err := cache.Fetch(startPage())
	if err != nil {
		return nil, fmt.Errorf("fetch start: %w", err)
	}
	klog.Infof("start page cached: %v cookies: %+v", sr.Cached, sr.Cookies)

	var results []Result

	for i := 0; i < maxPages; i++ {
		klog.Infof("page: %d", i)
		req := pageRequest(crit, date, i)
		req.Cookies = sr.Cookies

		resp, err := cache.Fetch(req)
		if err != nil {
			return nil, fmt.Errorf("fetch: %w", err)
		}

		var jr jaxResponse
		err = json.Unmarshal(resp.Body, &resp)
		if err != nil {
			return nil, fmt.Errorf("unmarshal: %w", err)
		}

		if jr.Control.CurrentPage != i {
			return nil, fmt.Errorf("got page %d, expected page %d. control: %+v", jr.Control.CurrentPage, i, jr.Control)
		}

		for _, jrs := range jr.Records {
			r := Result{
				Name: jrs.Name,
			}

			results = append(results, r)
		}

		if !resp.Cached {
			klog.V(1).Infof("Previous request was uncached, sleeping ...")
			time.Sleep(uncachedDelay)
		}

		if jr.Control.CurrentPage == jr.TotalPages {
			klog.Infof("fetched final page (%d)", jr.Control.CurrentPage)
			break
		}
	}

	return results, nil
}
