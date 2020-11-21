package backend

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/tstromberg/campwiz/pkg/cache"
	"github.com/tstromberg/campwiz/pkg/campwiz"
)

func TestRCPageRequest(t *testing.T) {
	date, err := time.Parse("2006-01-02", "2021-02-12")
	if err != nil {
		t.Fatalf("time parse: %v", err)
	}
	q := campwiz.Query{
		StayLength: 4,
		Lon:        -122.07237049999999,
		Lat:        37.4092297,
	}

	got, err := rcPageRequest(q, date)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	want := cache.Request{
		Method:      "POST",
		URL:         "https://calirdr.usedirect.com/rdr/rdr/search/place",
		Referrer:    "https://www.reservecalifornia.com",
		MaxAge:      time.Duration(6 * time.Hour),
		ContentType: "application/json",
		Body:        []byte{},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("rcPageRequest() mismatch (-want +got):\n%s", diff)
	}
}
