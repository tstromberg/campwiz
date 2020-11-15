package provider

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestParseSearchPage(t *testing.T) {
	bs, err := ioutil.ReadFile("testdata/search.json")
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

	got, gotPage, gotTotal, err := parseSearchPage(bs, date, q)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	if gotPage != 0 {
		t.Errorf("got page: %d, want: %d", gotPage, 0)
	}

	if gotTotal != 17 {
		t.Errorf("got total: %d, want: %d", gotTotal, 17)
	}

	want := []Result{
		{
			ID:       "STAN_1040013",
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
