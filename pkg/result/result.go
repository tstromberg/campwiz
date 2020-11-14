package result

import (
	"fmt"
	"time"
)

type MEntry struct {
	Key     string
	Name    string
	SRating int
	Desc    string
	Locale  string
}

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
	M            MEntry
}

// Returns a unique key for a specific site.
func (r Result) SiteKey() string {
	return fmt.Sprintf("%s:%s:%s", r.ContractCode, r.ParkId, r.Name)
}

type Results []Result

func (slice Results) Len() int {
	return len(slice)
}

func (slice Results) Less(i, j int) bool {
	return slice[i].M.SRating > slice[j].M.SRating
}

func (slice Results) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}
