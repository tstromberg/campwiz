package backend

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/tstromberg/campwiz/pkg/campwiz"
)

func TestEmptyParse(t *testing.T) {
	ra := &Empty{}

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

	got, err := ra.parse([]byte("hello"), date, q)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	want := []campwiz.Result{
		{
			Name:     "Empty Site",
			Distance: 99,
			Availability: []campwiz.Availability{
				{
					SiteKind: campwiz.Tent,
					Date:     time.Date(2021, 2, 12, 0, 0, 0, 0, time.UTC),
					URL:      "https://www.example.com/",
				},
			},
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("parseResp() mismatch (-want +got):\n%s", diff)
	}
}
