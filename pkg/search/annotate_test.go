package search

import (
	"testing"

	"github.com/tstromberg/campwiz/pkg/campwiz"
)

func TestFindBestMatch(t *testing.T) {
	props := map[string]*campwiz.Property{
		"/ca/chico/zlky": {
			ID:        "/ca/chico/zlky",
			Name:      "Zlky",
			URL:       "http://www.fs.usda.gov/elsewhere",
			ManagedBy: "Elk River National Forest",
			Campgrounds: []*campwiz.Campground{{
				ID:     "default",
				Name:   "Mt. Elky",
				ResURL: "http://www.recreation.gov",
				Refs: map[string]*campwiz.Ref{
					"cc": {
						Name:   "Elky",
						Desc:   "This Forest Service campground sits along an open plain.",
						Rating: 2,
						Locale: "Near Chico",
					},
				},
			}},
		},
		"/ca/campwiz": {
			ID:        "/ca/campwiz",
			Name:      "Campwiz National Forest",
			URL:       "http://www.fs.usda.gov/elsewhere",
			ManagedBy: "Thomas Stromberg",
			Campgrounds: []*campwiz.Campground{{
				ID:   "campy_left",
				Name: "Campy Left",
				Refs: map[string]*campwiz.Ref{
					"cc": {
						Name:   "Campy Left",
						Desc:   "This camp is out left field. a tiny, secluded, bug in a program. Bad? Of course, it’s bad.",
						Rating: 1,
						Locale: "on the Left Fork of the Test River in Campwiz National Forest",
					},
				},
			},
				{
					ID:   "campy_right",
					Name: "Campy Right",
					Refs: map[string]*campwiz.Ref{
						"cc": {
							Name:   "Campy Right",
							Desc:   "This camp is out left field. a tiny, secluded, bug in a program. Bad? Of course, it’s bad.",
							Rating: 9,
							Locale: "on the Right Fork of the Test River in Campwiz National Forest",
						},
					},
				},
			}},
	}

	var tests = []struct {
		in    string
		score int
		id    string
	}{
		{`Sad River`, NoMatch, ""},
		{`Campy Right`, NameMatch, "campy_right"},
		{`Just Campy Left`, SubNameMatch, "campy_left"},
		{`Mount Elky`, DoubleMangledNameMatch, "default"},
		{`Zlky`, SingleParkNameMatch, "default"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := findBestMatch(campwiz.Result{Name: tt.in}, props)
			if got.Score != tt.score {
				t.Errorf("got score %d %q, want %d %q: %+v", got.Score, scoreNames[got.Score], tt.score, scoreNames[tt.score], got)
			}

			if tt.id != "" {
				if got.Campground == nil {
					t.Fatalf("got campground nil, expected %q: %+v", tt.id, got)
				}
				if got.Campground.ID != tt.id {
					t.Fatalf("got campground %q, expected %q: %+v", got.Campground.ID, tt.id, got)
				}
			}
		})
	}

}
