package mangle

import (
	"strings"

	"github.com/tstromberg/campwiz/pkg/campwiz"
)

func SiteKind(title string, kind string, sid string) campwiz.SiteKind {
	if strings.Contains(sid, "RV") {
		return campwiz.RV
	}

	for _, in := range []string{sid, kind, title} {
		in = nonWordRe.ReplaceAllString(in, " ")
		for _, w := range strings.Split(strings.ToLower(in), " ") {
			switch w {
			case "ada", "accessible", "handicapped":
				if strings.Contains(kind, "RV") {
					return campwiz.RVADA
				}
				return campwiz.TentADA
			case "tent", "tent/non-electric":
				return campwiz.Tent
			case "horse", "equestrian":
				return campwiz.Equestrian
			case "RV", "hook-up", "hookup", "rv/electric":
				return campwiz.RV
			case "cabin", "yurt", "lodge", "hotel", "hostel", "motel", "lodging":
				return campwiz.Lodging
			case "boat", "boat-in", "kayak", "canoe":
				return campwiz.Boat
			case "group":
				return campwiz.Group
			case "day", "picnic":
				return campwiz.Day
			}
		}
	}

	return campwiz.Tent
}
