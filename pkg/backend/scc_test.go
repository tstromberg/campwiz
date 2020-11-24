package backend

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/tstromberg/campwiz/pkg/campwiz"
)

func TestSantaClaraCountyParse(t *testing.T) {
	b := &SantaClaraCounty{}

	date, err := time.Parse("2006-01-02", "2021-02-12")
	if err != nil {
		t.Fatalf("time parse: %v", err)
	}
	q := campwiz.Query{
		StayLength:  4,
		Lon:         -122.07237049999999,
		Lat:         37.4092297,
		MaxDistance: 100,
	}

	got, err := b.parse([]byte("hello"), date, q)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	want := []campwiz.Result{
		{
			ID:       "STAN_1040013",
			Name:     "FRANK RAINES REGIONAL PARK",
			Distance: 62.91,
			Availability: []campwiz.Availability{
				{URL: "https://www.reserveamerica.com/camping/frank-raines-regional-park/r/facilityDetails.do?contractCode=STAN&parkId=1040013&arrivalDate=2021-02-12&lengthOfStay=4"},
			},
		},
		{
			ID:       "PRCG_1060800",
			Name:     "Clear Lake Campground",
			Distance: 81.47,
			Availability: []campwiz.Availability{
				{URL: "https://www.reserveamerica.com/camping/frank-raines-regional-park/r/facilityDetails.do?contractCode=STAN&parkId=1040013&arrivalDate=2021-02-12&lengthOfStay=4"},
			},
		},
		{
			ID:       "STAN_1040012",
			Name:     "WOODWARD RESERVOIR REGIONAL PARK",
			Distance: 85.81,
			Availability: []campwiz.Availability{
				{URL: "https://www.reserveamerica.com/camping/frank-raines-regional-park/r/facilityDetails.do?contractCode=STAN&parkId=1040013&arrivalDate=2021-02-12&lengthOfStay=4"},
			},
		},
		{
			ID:       "STAN_1040011",
			Name:     "MODESTO RESERVOIR REGIONAL PARK",
			Distance: 98.04,
			Availability: []campwiz.Availability{
				{URL: "https://www.reserveamerica.com/camping/frank-raines-regional-park/r/facilityDetails.do?contractCode=STAN&parkId=1040013&arrivalDate=2021-02-12&lengthOfStay=4"},
			},
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("parseResp() mismatch (-want +got):\n%s", diff)
	}
}
