package campwiz

import (
	"time"
)

type Availability struct {
	SiteType string
	Count    int
	Date     time.Time
	URL      string
}

// Result is supposed to be a vendor neutral result of results
type Result struct {
	ResURL string
	ResID  string

	Name     string
	Distance float64

	Rating float64

	Desc string
	URL  string

	Availability []Availability
	Features     []string
	Locale       string

	KnownCampground Campground
}
