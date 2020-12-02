package backend

import (
	"encoding/json"
	"fmt"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"time"

	"github.com/tstromberg/campwiz/pkg/cache"
	"github.com/tstromberg/campwiz/pkg/campwiz"
	"k8s.io/klog/v2"
)

// RAmerica handles RAmerica queries
type RAmerica struct {
	store cache.Store
	jar   *cookiejar.Jar
}

// Name is a human readable name
func (b *RAmerica) Name() string {
	return "RAmerica"
}

// List lists available sites
func (b *RAmerica) List(q campwiz.Query) ([]campwiz.Result, error) {
	klog.Infof("RAmerica.List: %+v", q)
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
func (b *RAmerica) url(s string) string {
	return "https" + "://" + "www." + "reserve" + "america.com" + s
}

// req generates a search request
func (b *RAmerica) req(c campwiz.Query, arrival time.Time, num int) cache.Request {
	return cache.Request{
		URL:      b.url("/jaxrs-json/search"),
		Referrer: b.url("/"),
		Jar:      b.jar,
		Form: url.Values{
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
		},
	}
}

// startPage generates an initial page request
func (b *RAmerica) startPage() cache.Request {
	return cache.Request{URL: b.url("/explore/search-results"), Referrer: b.url("/"), Jar: b.jar}
}

type raControl struct {
	CurrentPage int
	PageSize    int
}

type raRecord struct {
	NamingID  string
	Name      string
	Proximity float64
	Details   raDetails
}

type raAvailability struct {
	Available      bool
	ReservableType string
}

type raDetails struct {
	BaseURL      string
	Availability raAvailability
}

type raResponse struct {
	TotalRecords int
	TotalPages   int
	StartIndex   int
	EndIndex     int
	Control      raControl
	Records      []raRecord
}

// parse parses the search response
func (b *RAmerica) parse(bs []byte, date time.Time, q campwiz.Query) ([]campwiz.Result, int, int, error) {
	klog.V(1).Infof("parsing %d bytes", len(bs))
	var jr raResponse
	err := json.Unmarshal(bs, &jr)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("unmarshal: %w", err)
	}
	klog.V(2).Infof("unmarshalled: %+v", jr)

	var results []campwiz.Result
	for _, r := range jr.Records {
		if q.MaxDistance > 0 && int(r.Proximity) > q.MaxDistance {
			klog.V(1).Infof("Skipping %s - too far (%.0f miles)", r.Name, r.Proximity)
			continue
		}

		if !r.Details.Availability.Available {
			continue
		}

		a := campwiz.Availability{
			SiteType: "campsite",
			Date:     date,
			URL:      b.url(r.Details.BaseURL + "&arrivalDate=" + date.Format("2006-01-02") + "&lengthOfStay=" + strconv.Itoa(q.StayLength)),
		}

		rr := campwiz.Result{
			ResURL:       b.url("/"),
			ResID:        r.NamingID,
			Name:         r.Name,
			Distance:     r.Proximity,
			Availability: []campwiz.Availability{a},
		}

		klog.Infof("%s is available: %+v", r.Name, rr)
		results = append(results, rr)
	}

	return results, jr.Control.CurrentPage, jr.TotalPages, nil
}

// avail lists sites available on a single date
func (b *RAmerica) avail(q campwiz.Query, d time.Time) ([]campwiz.Result, error) {
	var results []campwiz.Result

	for i := 0; i < maxPages; i++ {
		req := b.req(q, d, i)
		resp, err := cache.Fetch(req, b.store)
		if err != nil {
			return nil, fmt.Errorf("fetch: %w", err)
		}

		prs, currentPage, totalPages, err := b.parse(resp.Body, d, q)
		if err != nil {
			return nil, fmt.Errorf("parse: %w", err)
		}

		if currentPage != i {
			return nil, fmt.Errorf("got page %d, expected page %d", currentPage, i)
		}

		results = append(results, prs...)

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
