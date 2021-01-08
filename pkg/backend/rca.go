package backend

import (
	"encoding/json"
	"fmt"
	"net/http/cookiejar"
	"sort"
	"strings"
	"time"

	"github.com/tstromberg/campwiz/pkg/cache"
	"github.com/tstromberg/campwiz/pkg/campwiz"
	"github.com/tstromberg/campwiz/pkg/mangle"
	"k8s.io/klog/v2"
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

type availParams struct {
	CategoryID     int `json:"CategoryId"`
	ChooseActivity int `json:"ChooseActivity"`
	NoOfRecords    int `json:"NoOfRecords"`
	Page1          int `json:"Page1"`
	PageIndex      int `json:"PageIndex"`
	PageSize       int `json:"PageSize"`
	ParkCategory   int `json:"ParkCategory"`

	StartDate string `json:"StartDate"`
	Nights    string `json:"Nights"` // should be an int, but their interface uses a string
}

// googleParams is used by the /GetPlaceData endpoint (advanced search page)
type googleParams struct {
	Latitude  string `json:"Latitude"`
	Longitude string `json:"Longitude"`

	// Ignored: South, North, East, West
	Filter    bool `json:"Filter"`
	ZoomLevel int  `json:"ZoomLevel"`

	AvailabilitySearchParams availParams `json:"AvailabilitySearchParams"`
}

type rcAdvancedRequest struct {
	GooglePlaceSearchParameters googleParams `json:"googlePlaceSearchParameters"`
	ScreenResolution            int
}

// req creates the request object for a search.
func (b *RCaliforniaAdv) req(q campwiz.Query, arrival time.Time) (cache.Request, error) {
	rcr := rcAdvancedRequest{
		GooglePlaceSearchParameters: googleParams{
			Latitude:  fmt.Sprintf("%.4f", q.Lat),
			Longitude: fmt.Sprintf("%.4f", q.Lon),
			ZoomLevel: 6,
			AvailabilitySearchParams: availParams{
				StartDate: arrival.Format("01-02-2006"),
				Nights:    fmt.Sprintf("%d", q.StayLength),
			},
		},
		ScreenResolution: 1422,
	}

	body, err := json.Marshal(&rcr)
	if err != nil {
		return cache.Request{}, fmt.Errorf("marshal: %w", err)
	}

	r := cache.Request{
		Method:      "POST",
		URL:         "https://www.reservecalifornia.com/CaliforniaWebHome/Facilities/AdvanceSearch.aspx/GetPlaceData",
		Referrer:    "https://www.reservecalifornia.com/CaliforniaWebHome/Facilities/AdvanceSearch.aspx",
		MaxAge:      searchPageExpiry,
		ContentType: "application/json",
		Body:        body,
	}

	return r, nil
}

type spot struct {
	Name  string `json:"FacilityName"`
	Type  string `json:"SpottypeName"`
	Count int    `json:"SpotCount"`
}

type facilityInfo struct {
	Name      string  `json:"FacilityName"`
	Latitude  float64 `json:"FacilityBoundryLatitude"`
	Longitude float64 `json:"FacilityBoundryLongitude"`
	Spots     []spot  `json:"JsonFacilitySpots"`
}

type placeInfo struct {
	Name         string         `json:"DisplayName"`
	Distance     int            `json:"PlaceDistance"`
	Description  string         `json:"FullDescription"`
	Latitude     float64        `json:"Latitude"`
	Longitude    float64        `json:"Longitude"`
	PlaceID      int            `json:"PlaceId"`
	ImageURL     string         `json:"ImageUrl"`
	Highlights   string         `json:"AllHightlights"` // that's really the field name: it's not a typo
	FacilityInfo []facilityInfo `json:"JsonFacilityInfos"`
	Available    bool           `json:"IsavailableSpots"`
	URL          string         `json:"PlaceinfoUrl"`
}

type rcaResponse struct {
	Data []placeInfo `json:"d"`
}

func (b *RCaliforniaAdv) parse(bs []byte, date time.Time, q campwiz.Query) ([]campwiz.Result, error) {

	var rr rcaResponse
	err := json.Unmarshal(bs, &rr)
	if err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	klog.V(2).Infof("unmarshalled data: %+v", rr)

	var results []campwiz.Result
	for _, p := range rr.Data {
		klog.Infof("found place: %+v", p)

		if !p.Available {
			continue
		}

		r := campwiz.Result{
			ResURL:       b.url("/"),
			ResID:        fmt.Sprintf("%2d", p.PlaceID),
			Name:         p.Name,
			Desc:         p.Description,
			URL:          p.URL,
			Features:     mangle.Features(p.Highlights),
			Distance:     float64(p.Distance),
			ImageURL:     p.ImageURL,
			Availability: []campwiz.Availability{},
		}

		klog.Infof("%s may be available: %+v", p.Name, rr)
		avail := map[string]*campwiz.Availability{}

		for _, fi := range p.FacilityInfo {
			for _, sp := range fi.Spots {
				kind := mangle.SiteKind(sp.Name, sp.Type, "")
				name := strings.TrimSpace(strings.Split(sp.Name, "(")[0])
				key := fmt.Sprintf("%s=%s", name, kind)
				a, ok := avail[key]
				if ok {
					a.SpotCount++
					continue
				}

				avail[key] = &campwiz.Availability{
					Kind:      kind,
					Name:      name,
					Desc:      sp.Type,
					SpotCount: sp.Count,
					Date:      date,
				}
			}
		}

		for _, a := range avail {
			r.Availability = append(r.Availability, *a)
		}

		sort.Slice(r.Availability, func(i, j int) bool {
			return string(r.Availability[i].Kind)+r.Availability[i].Desc < string(r.Availability[j].Kind)+r.Availability[j].Desc
		})
		results = append(results, r)
	}

	return results, nil
}

// avail returns sites available on a single date
func (b *RCaliforniaAdv) avail(q campwiz.Query, d time.Time) ([]campwiz.Result, error) {
	return nil, nil
}
