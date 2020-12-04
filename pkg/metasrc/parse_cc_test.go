package metasrc

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tstromberg/campwiz/pkg/campwiz"
)

func TestHTMLText(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{`&#8220;Lions.&#8221;`, `“Lions.”`},
		{`winter <a id="page_499"></a>weekends.`, "winter weekends."},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := htmlText(tt.in)
			if got != tt.out {
				t.Errorf("got %q, want %q", got, tt.out)
			}
		})
	}
}

func TestCCPropertyKey(t *testing.T) {
	tests := []struct {
		name   string
		locale string
		out    string
	}{
		{`Sad River`, "in Elk River National Forest", "/ca/elk_river"},
		{`Sad River`, "North of Los Angeles", "/ca/los_angeles/sad_river"},
		{`Sam's Place`, "Barstow", "/ca/barstow/sams_place"},
	}

	for _, tt := range tests {
		t.Run(tt.out, func(t *testing.T) {
			got := ccPropertyKey(tt.name, tt.locale)
			if got != tt.out {
				t.Errorf("got %q, want %q", got, tt.out)
			}
		})
	}
}

func TestCC(t *testing.T) {
	f, err := os.Open("testdata/cc.html")
	if err != nil {
		t.Fatalf("readfile: %v", err)
	}

	got := map[string]*campwiz.Property{}

	if err := CC(f, got); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	want := map[string]*campwiz.Property{
		"/ca/chico/elky": {
			ID:        "/ca/chico/elky",
			Name:      "Elky",
			URL:       "http://www.fs.usda.gov/elsewhere",
			ManagedBy: "Elk River National Forest",
			Campgrounds: []*campwiz.Campground{{
				ID:     "default",
				Name:   "Elky",
				ResURL: "http://www.recreation.gov",
				Refs: map[string]*campwiz.Ref{
					"cc": {
						Name:    "Elky",
						Desc:    "This Forest Service campground sits along an open plain.",
						Contact: "Elk River National Forest, Banana Peel Ranger District, 107/374-1234, www.fs.usda.gov/elsewhere.",
						Rating:  2,
						Locale:  "Near Chico",
					},
				},
			}},
		},
		"/ca/campwiz": {
			ID:        "/ca/campwiz",
			Name:      "Campwiz National Forest",
			URL:       "http://www.fs.usda.gov/elsewhere",
			ManagedBy: "Thomas Stromberg",
			Campgrounds: []*campwiz.Campground{
				{
					ID:   "campy_left",
					Name: "Campy Left",
					Refs: map[string]*campwiz.Ref{
						"cc": {
							Name:    "Campy Left",
							Desc:    "This camp is out left field. a tiny, secluded, bug in a program. Bad? Of course, it’s bad.",
							Contact: "Thomas Stromberg, 405/487-5555, www.fs.usda.gov/elsewhere.",
							Rating:  1,
							Locale:  "on the Left Fork of the Test River in Campwiz National Forest",
						},
					},
				},
				{
					ID:   "campy_right",
					Name: "Campy Right",
					Refs: map[string]*campwiz.Ref{
						"cc": {
							Name:    "Campy Right",
							Desc:    "This camp is out left field. a tiny, secluded, bug in a program. Bad? Of course, it’s bad.",
							Contact: "Thomas Stromberg, 405/487-5555, www.fs.usda.gov/elsewhere.",
							Rating:  9,
							Locale:  "on the Right Fork of the Test River in Campwiz National Forest",
						},
					},
				},
			},
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("CC() mismatch (-want +got):\n%s\nRAW: %+v", diff, got)
	}
}

func TestBestCC(t *testing.T) {
	f, err := os.Open("testdata/best_cc.html")
	if err != nil {
		t.Fatalf("readfile: %v", err)
	}

	got := map[string]*campwiz.Property{}

	if err := CC(f, got); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	want := map[string]*campwiz.Property{
		"/ca/patalinx_planet/smal_harbor": {
			ID:   "/ca/patalinx_planet/smal_harbor",
			Name: "Smal Harbor",
			Campgrounds: []*campwiz.Campground{
				{
					ID:   "default",
					Name: "Smal Harbor",
					Refs: map[string]*campwiz.Ref{
						"cc": {
							Name:   "Smal Harbor",
							Desc:   "There is plenty to do here!",
							Locale: "on Patalinx Planet",
							Rating: 10,
							Lists: []campwiz.RefList{{
								Title: "Best Planet Retreats",
								Place: 8,
							}},
						},
					},
				},
			},
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("CC() mismatch (-want +got):\n%s\nRAW: %+v", diff, got)
	}
}
