package search

import (
	"io/ioutil"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/tstromberg/campwiz/pkg/cache"
)

func TestParseSMCSearchPage(t *testing.T) {
	bs, err := ioutil.ReadFile("testmetadata/smc_feed.xml")
	if err != nil {
		t.Fatalf("readfile: %v", err)
	}
	date, err := time.Parse("2006-01-02", "2021-02-12")
	if err != nil {
		t.Fatalf("time parse: %v", err)
	}
	q := Query{
		StayLength: 4,
		Lon:        -122.07237049999999,
		Lat:        37.4092297,
	}

	got, err := parseSMCSearchPage("coyote-point", bs, date, q)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	want := []Result{
		{
			ID:       "parks.smcgov.org/coyote-point",
			Name:     "FRANK RAINES REGIONAL PARK",
			Distance: 62.91,
			Availability: []Availability{
				{URL: "https://www.reserveamerica.com/camping/frank-raines-regional-park/r/facilityDetails.do?contractCode=STAN&parkId=1040013&arrivalDate=2021-02-12&lengthOfStay=4"},
			},
		},
		{
			ID:       "PRCG_1060800",
			Name:     "Clear Lake Campground",
			Distance: 81.47,
			Availability: []Availability{
				{URL: "https://www.reserveamerica.com/camping/frank-raines-regional-park/r/facilityDetails.do?contractCode=STAN&parkId=1040013&arrivalDate=2021-02-12&lengthOfStay=4"},
			},
		},
		{
			ID:       "STAN_1040012",
			Name:     "WOODWARD RESERVOIR REGIONAL PARK",
			Distance: 85.81,
			Availability: []Availability{
				{URL: "https://www.reserveamerica.com/camping/frank-raines-regional-park/r/facilityDetails.do?contractCode=STAN&parkId=1040013&arrivalDate=2021-02-12&lengthOfStay=4"},
			},
		},
		{
			ID:       "STAN_1040011",
			Name:     "MODESTO RESERVOIR REGIONAL PARK",
			Distance: 98.04,
			Availability: []Availability{
				{URL: "https://www.reserveamerica.com/camping/frank-raines-regional-park/r/facilityDetails.do?contractCode=STAN&parkId=1040013&arrivalDate=2021-02-12&lengthOfStay=4"},
			},
		},
		{
			ID:       "PRCG_1073051",
			Name:     "Yosemite Ridge Resort",
			Distance: 130.33,
			Availability: []Availability{
				{URL: "https://www.reserveamerica.com/camping/frank-raines-regional-park/r/facilityDetails.do?contractCode=STAN&parkId=1040013&arrivalDate=2021-02-12&lengthOfStay=4"},
			},
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("parseSearchPage() mismatch (-want +got):\n%s", diff)
	}
}

func TestSMCSiteRequest(t *testing.T) {
	date, err := time.Parse("2006-01-02", "2021-02-12")
	if err != nil {
		t.Fatalf("time parse: %v", err)
	}
	q := Query{
		StayLength: 4,
		Lon:        -122.07237049999999,
		Lat:        37.4092297,
	}

	got := smcSiteRequest("coyote-point", q, date)

	want := cache.Request{
		Method:   "GET",
		URL:      "https://secure.itinio.com/sanmateo/feed.html",
		Referrer: "https://https://secure.itinio.com/sanmateo/coyote-point",
		MaxAge:   time.Duration(6 * time.Hour),
		Body:     nil,
		Form: url.Values{
			"code":      {"0.6046602879796196"},
			"endDate":   {"2021-02-16"},
			"startDate": {"2021-02-12"},
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("rcPageRequest() mismatch (-want +got):\n%s", diff)
	}
}
