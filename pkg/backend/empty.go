package backend

import (
	"fmt"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"time"

	"github.com/tstromberg/campwiz/pkg/cache"
	"github.com/tstromberg/campwiz/pkg/campwiz"
	"k8s.io/klog/v2"
)

// Empty handles Empty queries
type Empty struct {
	store cache.Store
	jar   *cookiejar.Jar
}

// Name is a human readable name
func (b *Empty) Name() string {
	return "Empty"
}

// List lists available sites
func (b *Empty) List(q campwiz.Query) ([]campwiz.Result, error) {
	klog.Infof("Empty.List: %+v", q)
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
func (b *Empty) url(s string) string {
	return "https://www.example.com/"
}

// req generates a search request
func (b *Empty) req(c campwiz.Query, arrival time.Time) cache.Request {
	return cache.Request{
		URL:      b.url("/search"),
		Referrer: b.url("/"),
		Jar:      b.jar,
		Form: url.Values{
			"lng": {fmt.Sprintf("%3.3f", c.Lon)},  // Longitude
			"lat": {fmt.Sprintf("%3.3f", c.Lat)},  // Latitude
			"arv": {arrival.Format("2006-01-02")}, // arrival date,
			"lsy": {strconv.Itoa(c.StayLength)},   // length of stay
		},
	}
}

// startPage generates an initial page request
func (b *Empty) startPage() cache.Request {
	return cache.Request{URL: b.url("/start"), Referrer: b.url("/"), Jar: b.jar}
}

// parse parses the search response
func (b *Empty) parse(bs []byte, date time.Time, q campwiz.Query) ([]campwiz.Result, error) {
	rr := campwiz.Result{
		ResID:    "EMPTY_1",
		Name:     "Empty Site",
		Distance: float64(q.MaxDistance) - 1,
		Availability: []campwiz.Availability{
			{
				Kind: campwiz.Tent,
				Date: date,
				URL:  b.url("/site" + "&arrivalDate=" + date.Format("2006-01-02") + "&lengthOfStay=" + strconv.Itoa(q.StayLength)),
			},
		},
	}
	return []campwiz.Result{rr}, nil
}

// avail lists sites available on a single date
func (b *Empty) avail(q campwiz.Query, d time.Time) ([]campwiz.Result, error) {
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
