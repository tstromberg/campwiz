package search

import (
	"testing"
)

func TestFilter(t *testing.T) {
	/* TODO: refactor for new result struct
	as := []campwiz.AnnotatedResult{
		{
			Result: campwiz.Result{
				Name:     "pretty close",
				Distance: 30.45,
			},
			Refs: []campwiz.Campground{
				{
					Source: campwiz.RefSource{
						RatingDesc: "Scenery",
					},
					Rating: 7.0,
					Desc:   "Tucked into a redwood forest",
				},
			},
		},
		{
			Result: campwiz.Result{
				Name:     "ugly far",
				Distance: 90.45,
			},
			Refs: []campwiz.Campground{
				{
					Source: campwiz.RefSource{
						RatingDesc: "Scenery",
					},
					Rating: 2.0,
					Desc:   "Hidden in an abandoned dump",
				},
			},
		},
	}

	var tests = []struct {
		in  campwiz.Query
		out []string
	}{
		{campwiz.Query{MaxDistance: 35.0}, []string{"pretty close"}},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := filter(tt.in)
			gotNames := []string{}
			for _, r := range got {
				gotNames = append(gotNames, got.Result.Name)
			}

			if gotNames != tt.out {
				t.Errorf("got %q, want %q", gotNames, tt.out)
			}
		})
	}
	*/
}
