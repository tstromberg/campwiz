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
	ID string

	Name     string
	Distance float64

	Desc string
	URL  string

	Availability []Availability
	Features     []string
}

// AnnotatedResult is a result with associated cross-reference data
type AnnotatedResult struct {
	// A mix of the best available data
	Name     string
	Distance float64
	Desc     string
	Features []string
	Locale   string

	// Original data
	Result Result
	Refs   []Ref
}
