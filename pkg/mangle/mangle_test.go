package mangle

import "testing"

func TestLocalizer(t *testing.T) {
	var tests = []struct {
		in  string
		out string
	}{
		{"on the Shasta River", "Shasta River"},
		{"in Shasta-Trinity Park", "Shasta-Trinity Park"},
		{"on the Yellow River in Shasta-Trinity Park", "Yellow River / Shasta-Trinity Park"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := Localizer(tt.in)
			if got != tt.out {
				t.Errorf("got %q, want %q", got, tt.out)
			}
		})
	}
}

func TestShortest(t *testing.T) {
	var tests = []struct {
		in  string
		out string
	}{
		{"Shasta River National Forest", "Shasta River"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := Shortest(tt.in)
			if got != tt.out {
				t.Errorf("got %q, want %q", got, tt.out)
			}
		})
	}
}
