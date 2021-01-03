package mangle

import (
	"strings"

	"github.com/tstromberg/campwiz/pkg/campwiz"
)

func SiteKind(title string, kind string, sid string) campwiz.SiteKind {
	if strings.Contains(sid, "RV") {
		return campwiz.RV
	}
	if strings.Contains(title, "Picnic") {
		return campwiz.Day
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
			case "tent":
				return campwiz.Tent
			case "horse", "equestrian":
				return campwiz.Equestrian
			case "RV", "hook", "hookup", "electric":
				return campwiz.RV
			case "cabin", "yurt", "lodge", "hotel", "hostel", "motel", "lodging":
				return campwiz.Lodging
			case "boat", "kayak", "canoe":
				return campwiz.Boat
				// before group
			case "day", "picnic":
				return campwiz.Day
			case "group":
				return campwiz.Group
			case "walk", "hike":
				return campwiz.Walk
			}
		}
	}

	return campwiz.Tent
}
