package query

import (
	"fmt"
	"time"
)


type Availability struct {
	Date       time.Time
	Nights     int
	Group      int64
	Standard   int64
	Accessible int64
	Equestrian int64
	Day        int64
	Boat       int64
	WalkIn     int64
	Rv         int64
}

type Result struct {
	Name         string
	ContractCode string
	ParkId       string
	Distance     float64
	State        string
	ShortDesc    string
	Availability []Availability
	URL          string
	Amenities    string
	M            Xref
}

// Returns a unique key for a specific site.
func (r Result) SiteKey() string {
	return fmt.Sprintf("%s:%s:%s", r.ContractCode, r.ParkId, r.Name)
}
