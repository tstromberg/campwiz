package mangle

import "testing"

func TestLocale(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{"on the Shasta River", "Shasta River"},
		{"in Shasta-Trinity Park", "Shasta-Trinity Park"},
		{"on the Yellow River in Shasta-Trinity Park", "Yellow River / Shasta-Trinity Park"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := Locale(tt.in)
			if got != tt.out {
				t.Errorf("got %q, want %q", got, tt.out)
			}
		})
	}
}

func TestLocaleProperty(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{"on the Shasta River", ""},
		{"in Shasta-Trinity Park", "Shasta-Trinity Park"},
		{"in the Shasta Wilderness Area", "Shasta Wilderness Area"},
		{"on the Yellow River / Shasta-Trinity Park", "Shasta-Trinity Park"},
		{"in Zonk Regional Park near Modesto", "Zonk Regional Park"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := LocaleProperty(tt.in)
			if got != tt.out {
				t.Errorf("got %q, want %q", got, tt.out)
			}
		})
	}
}

func TestShortLocale(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{"on the Shasta River", "Shasta River"},
		{"in Shasta-Trinity Park", "Shasta-Trinity"},
		{"on the Yellow River in Shasta-Trinity Park", "Shasta-Trinity"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := ShortLocale(tt.in)
			if got != tt.out {
				t.Errorf("got %q, want %q", got, tt.out)
			}
		})
	}
}

func TestShortest(t *testing.T) {
	tests := []struct {
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

func TestTitle(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{"BOW WILLOW AT MOUNTAIN PALM SPRINGS", "Bow Willow at Mountain Palm Springs"},
		{"Shasta River National Forest", "Shasta River National Forest"},
		{"my RV Park", "My RV Park"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := Title(tt.in)
			if got != tt.out {
				t.Errorf("got %q, want %q", got, tt.out)
			}
		})
	}
}
