package campwiz

import "time"

const (
	// Constants from https://developer.active.com/docs/read/Campground_Search_API
	// Site Types

	RV      = 2001  // RV website
	Cabin   = 10001 // Cabin or lodging
	Tent    = 2003
	Trailer = 2002
	Group   = 9002
	Day     = 9001
	Horse   = 3001
	Boat    = 2004

	// Features

	Biking                 = 4001
	Boating                = 4002
	EquipmentRental        = 4003
	Fishing                = 4004
	Golf                   = 4005
	Hiking                 = 4006
	HorsebackRiding        = 4007
	Hunting                = 4008
	RecreationalActivities = 4009
	ScenicTrails           = 4010
	Sports                 = 4011
	Beach                  = 4012
	Winter                 = 4013
)

// Query defines a list of attributes that can be sent to the camp engines
type Query struct {
	Lat         float64
	Lon         float64
	Dates       []time.Time
	StayLength  int
	MaxDistance int

	SiteTypes []int
	Features  []int
}
