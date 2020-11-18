package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/peterbourgon/diskv"
	"github.com/tstromberg/campwiz/pkg/cache"
	"k8s.io/klog/v2"
)

var (
	// raURL is the search URL to request reservation information from.
	raURL = "https://" + "www." + "reserve" + "america.com"

	// searchPageExpiry is how long search pages can be cached for.
	searchPageExpiry = time.Duration(6*3600) * time.Second

	// amount of time to sleep between uncached fetches
	uncachedDelay = time.Millisecond * 600

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
	Proximity float64
	Details   jaxDetails
}

type jaxAvailability struct {
	Available      bool
	ReservableType string
}

type jaxDetails struct {
	BaseURL      string
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

func mergeCookies(old []*http.Cookie, new []*http.Cookie) []*http.Cookie {
	merged := []*http.Cookie{}
	seen := map[string]bool{}

	for _, c := range new {
		merged = append(merged, c)
		seen[c.Name] = true
	}
	for _, c := range old {
		if !seen[c.Name] {
			merged = append(merged, c)
		}
	}

	return merged
}

func parseSearchPage(bs []byte, date time.Time, q Query) ([]Result, int, int, error) {
	var jr jaxResponse
	err := json.Unmarshal(bs, &jr)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("unmarshal: %w", err)
	}

	klog.V(2).Infof("unmarshalled data: %+v", jr)

	var results []Result
	for _, r := range jr.Records {
		if !r.Details.Availability.Available {
			continue
		}
		a := Availability{
			SiteType: "campsite",
			Date:     date,
			URL:      raURL + r.Details.BaseURL + "&arrivalDate=" + date.Format("2006-01-02") + "&lengthOfStay=" + strconv.Itoa(q.StayLength),
		}

		rr := Result{
			ID:           r.NamingID,
			Name:         r.Name,
			Distance:     r.Proximity,
			Availability: []Availability{a},
		}

		klog.Infof("%s is available: %+v", r.Name, rr)
		results = append(results, rr)
	}

	return results, jr.Control.CurrentPage, jr.TotalPages, nil
}

// searchRA runs a search for a single date
func searchRA(q Query, date time.Time, dv *diskv.Diskv) ([]Result, error) {
	klog.Infof("searchRA: %+v", q)

	// grab the current cookies
	sr, err := cache.Fetch(startPage(), dv)
	if err != nil {
		return nil, fmt.Errorf("fetch start: %w", err)
	}

	cookies := sr.Cookies
	klog.Infof("start page cached: %v cookies: %+v", sr.Cached, sr.Cookies)

	var results []Result

	for i := 0; i < maxPages; i++ {
		klog.Infof(">>>>>>>>> requesting page: %d", i)
		req := pageRequest(q, date, i)
		req.Cookies = cookies

		resp, err := cache.Fetch(req, dv)
		if err != nil {
			return nil, fmt.Errorf("fetch: %w", err)
		}
		cookies = mergeCookies(cookies, resp.Cookies)

		pageResults, currentPage, totalPages, err := parseSearchPage(resp.Body, date, q)
		if err != nil {
			return nil, fmt.Errorf("parse: %w", err)
		}

		if currentPage != i {
			return nil, fmt.Errorf("got page %d, expected page %d", currentPage, i)
		}

		results = append(results, pageResults...)

		if currentPage >= totalPages-1 {
			break
		}

		if !resp.Cached {
			klog.V(1).Infof("Previous request was uncached, sleeping ...")
			time.Sleep(uncachedDelay)
		}

	}

	klog.Infof("returning %d results", len(results))
	return results, nil
}
