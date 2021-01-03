package backend

import (
	"fmt"
	"net/http/cookiejar"
	"time"

	"github.com/tstromberg/campwiz/pkg/cache"
	"github.com/tstromberg/campwiz/pkg/campwiz"
)

// RCaliforniaAdv handles RCaliforniaAdv queries
type RCaliforniaAdv struct {
	store cache.Store
	jar   *cookiejar.Jar
}

// Name is a human readable name
func (b *RCaliforniaAdv) Name() string {
	return "RCaliforniaAdv"
}

// List lists available sites
func (b *RCaliforniaAdv) List(q campwiz.Query) ([]campwiz.Result, error) {
	var res []campwiz.Result
	for _, d := range q.Dates {
		rs, err := b.avail(q, d)
		if err != nil {
			return res, fmt.Errorf("onDate: %w", err)
		}
		res = append(res, rs...)
	}

	return mergeDates(res), nil
}

// url is the root URL to use for requests
func (b *RCaliforniaAdv) url(s string) string {
	return "https://" + "www." + "reserve" + "california.com" + s
}

// req creates the request object for a search.
func (b *RCaliforniaAdv) req(q campwiz.Query, arrival time.Time) (cache.Request, error) {
	return cache.Request{}, nil
}

type availParams struct {
	CategoryID     int `json:"CategoryId"`
	ChooseActivity int `json:"ChooseActivity"`
	NoOfRecords    int `json:"NoOfRecords"`
	Page1          int `json:"Page1"`
	PageIndex      int `json:"PageIndex"`
	PageSize       int `json:"PageSize"`
	ParkCategory   int `json:"ParkCategory"`

	StartDate string `json:"StartDate`
	Nights    string `json:"Nights` // should be an int, but their interface uses a string
}

// googleParams is used by the /GetPlaceData endpoint (advanced search page)
type googleParams struct {
	Latitude  string `json:"Latitude"`
	Longitude string `json:"Longitude"`

	// Ignored: South, North, East, West
	Filter    bool `json:"Filter"`
	ZoomLevel int  `json:"ZoomLevel"`

	AvailabilitySearchParams availParams `json:"AvailabilitySearchParams`
}

type rcAdvancedRequest struct {
	GooglePlaceSearchParameters googleParams `json:"googlePlaceSearchParameters"`
	ScreenResolution            int
}

func (b *RCaliforniaAdv) parse(bs []byte, date time.Time, q campwiz.Query) ([]campwiz.Result, error) {
	return nil, nil
}

// avail returns sites available on a single date
func (b *RCaliforniaAdv) avail(q campwiz.Query, d time.Time) ([]campwiz.Result, error) {
	return nil, nil
}
