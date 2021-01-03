package backend

import (
	"io/ioutil"
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

	bs, err := ioutil.ReadFile("testdata/scc.html")
	if err != nil {
		t.Fatalf("readfile: %v", err)
	}

	got, err := b.parse([]byte(bs), date, q)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	want := []campwiz.Result{
		{
			ResURL:   "https://gooutsideandplay.org/",
			ResID:    "coyote_lake",
			Name:     "Coyote Lake",
			Distance: 24.04395390049703,
			Availability: []campwiz.Availability{
				{
					SiteKind:  campwiz.TentADA,
					SiteDesc:  "Camping - Tent/Non-Electric",
					Date:      time.Date(2021, 0o2, 12, 0, 0, 0, 0, time.UTC),
					SpotCount: 1,
					Name:      "Coyote Lake",
					URL:       "https://gooutsideandplay.org/reservations/SiteDetails.asp?arrivedate=02/12/2021&departdate=2/16/2021&SiteID=184",
				},
				{
					SiteKind:  campwiz.RVADA,
					SiteDesc:  "Camping - RV/Electric",
					Date:      time.Date(2021, 0o2, 12, 0, 0, 0, 0, time.UTC),
					SpotCount: 1,
					Name:      "Coyote Lake",
					URL:       "https://gooutsideandplay.org/reservations/SiteDetails.asp?arrivedate=02/12/2021&departdate=2/16/2021&SiteID=102510",
				},
				{
					SiteKind:  campwiz.Tent,
					SiteDesc:  "Camping - Tent/Non-Electric",
					Date:      time.Date(2021, 0o2, 12, 0, 0, 0, 0, time.UTC),
					SpotCount: 52,
					Name:      "Coyote Lake",
					URL:       "https://gooutsideandplay.org/reservations/SiteDetails.asp?arrivedate=02/12/2021&departdate=2/16/2021&SiteID=134",
				},
				{
					SiteKind:  campwiz.RV,
					SiteDesc:  "Camping - RV/Electric",
					Date:      time.Date(2021, 0o2, 12, 0, 0, 0, 0, time.UTC),
					SpotCount: 10,
					Name:      "Coyote Lake",
					URL:       "https://gooutsideandplay.org/reservations/SiteDetails.asp?arrivedate=02/12/2021&departdate=2/16/2021&SiteID=121",
				},
			},
		},
		{
			ResURL:   "https://gooutsideandplay.org/",
			ResID:    "joseph_grant_park",
			Name:     "Joseph Grant Park",
			Distance: 24.04395390049703,
			Availability: []campwiz.Availability{
				{
					SiteKind:  "⛺",
					SiteDesc:  "Camping - Tent/Non-Electric",
					Name:      "Joseph Grant Park",
					SpotCount: 15,
					Date:      time.Date(2021, 0o2, 12, 0, 0, 0, 0, time.UTC),
					URL:       "https://gooutsideandplay.org/reservations/SiteDetails.asp?arrivedate=02/12/2021&departdate=2/16/2021&SiteID=201",
				},
				{
					SiteKind:  "🏇",
					SiteDesc:  "Camping - Tent/Non-Electric",
					Name:      "Joseph Grant Park",
					SpotCount: 7,
					Date:      time.Date(2021, 0o2, 12, 0, 0, 0, 0, time.UTC),
					URL:       "https://gooutsideandplay.org/reservations/SiteDetails.asp?arrivedate=02/12/2021&departdate=2/16/2021&SiteID=241",
				},
			},
		},
		{
			ResURL:   "https://gooutsideandplay.org/",
			ResID:    "mt_madonna_park",
			Name:     "Mt Madonna Park",
			Distance: 24.04395390049703,
			Availability: []campwiz.Availability{
				{
					SiteKind:  "♿⛺",
					SiteDesc:  "Camping - Tent/Non-Electric",
					Name:      "Mt Madonna Park",
					SpotCount: 1,
					Date:      time.Date(2021, 0o2, 12, 0, 0, 0, 0, time.UTC),
					URL:       "https://gooutsideandplay.org/reservations/SiteDetails.asp?arrivedate=02/12/2021&departdate=2/16/2021&SiteID=338",
				},
				{
					SiteKind:  "⛺",
					SiteDesc:  "Camping - Tent/Non-Electric",
					Name:      "Mt Madonna Park",
					SpotCount: 33,
					Date:      time.Date(2021, 0o2, 12, 0, 0, 0, 0, time.UTC),
					URL:       "https://gooutsideandplay.org/reservations/SiteDetails.asp?arrivedate=02/12/2021&departdate=2/16/2021&SiteID=292",
				},
				{
					SiteKind:  "🚙",
					SiteDesc:  "Camping - RV/Electric",
					Name:      "Mt Madonna Park",
					SpotCount: 5,
					Date:      time.Date(2021, 0o2, 12, 0, 0, 0, 0, time.UTC),
					URL:       "https://gooutsideandplay.org/reservations/SiteDetails.asp?arrivedate=02/12/2021&departdate=2/16/2021&SiteID=375",
				},
			},
		},
		{
			ResURL:   "https://gooutsideandplay.org/",
			ResID:    "sanborn",
			Name:     "Sanborn",
			Distance: 24.04395390049703,
			Availability: []campwiz.Availability{
				{
					SiteKind:  "🚙",
					SiteDesc:  "Camping - RV/Electric",
					Name:      "Sanborn",
					SpotCount: 7,
					Date:      time.Date(2021, 0o2, 12, 0, 0, 0, 0, time.UTC),
					URL:       "https://gooutsideandplay.org/reservations/SiteDetails.asp?arrivedate=02/12/2021&departdate=2/16/2021&SiteID=404",
				},
			},
		},
		{
			ResURL:   "https://gooutsideandplay.org/",
			ResID:    "uvas_canyon_park",
			Name:     "Uvas Canyon Park",
			Distance: 24.04395390049703,
			Availability: []campwiz.Availability{
				{
					SiteKind:  "♿⛺",
					SiteDesc:  "Camping - Tent/Non-Electric",
					Name:      "Uvas Canyon Park",
					SpotCount: 1,
					Date:      time.Date(2021, 0o2, 12, 0, 0, 0, 0, time.UTC),
					URL:       "https://gooutsideandplay.org/reservations/SiteDetails.asp?arrivedate=02/12/2021&departdate=2/16/2021&SiteID=477",
				},
				{
					SiteKind:  "⛺",
					SiteDesc:  "Camping - Tent/Non-Electric",
					Name:      "Uvas Canyon Park",
					SpotCount: 20,
					Date:      time.Date(2021, 0o2, 12, 0, 0, 0, 0, time.UTC),
					URL:       "https://gooutsideandplay.org/reservations/SiteDetails.asp?arrivedate=02/12/2021&departdate=2/16/2021&SiteID=468",
				},
			},
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("parseResp() mismatch (-want +got):\n%s\nraw: %+v\n", diff, got)
	}
}
