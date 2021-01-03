package campwiz

import (
	"time"
)

// Availability represents what's actually available at the site
type Availability struct {
	SiteKind SiteKind
	SiteDesc string

	Name        string
	Description string

	SpotCount int

	Date time.Time
	URL  string
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

	ImageURL string

	Availability []Availability
	Features     []string
	Locale       string

	KnownCampground *Campground
}
