package mangle

import (
	"testing"

	"github.com/tstromberg/campwiz/pkg/campwiz"
)

func TestSiteKind(t *testing.T) {
	tests := []struct {
		title string
		kind  string
		sid   string
		out   campwiz.Kind
	}{
		{"Frank Valley Horse Camp", "", "", campwiz.Equestrian},
		{"Shasta-Trinity Park", "", "", campwiz.Tent},
		{"Shasta-Trinity Boat-In", "", "", campwiz.Boat},
		{"Angel Island Group Campsite", "", "", campwiz.Group},
		{"Joseph Grant Park", "Camping - Tent/Non-Electric", "#8-Horse Camp Only *", campwiz.Equestrian},
		{"Moo", "?", "13RV", campwiz.RV},
		{"Zoo", "Camping - RV/Electric", "1E ADA", campwiz.AccessibleRV},
		{"Ayala Cove Group Picnic Area", "", "614", campwiz.Day},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			got := SiteKind(tt.title, tt.kind, tt.sid)
			if got != tt.out {
				t.Errorf("got %q, want %q", got, tt.out)
			}
		})
	}
}
