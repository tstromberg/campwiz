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

	Description string
	URL         string

	States []string

	Availability []Availability
	Features     string
}

// AnnotatedResult is a result with associated cross-reference data
type AnnotatedResult struct {
	Result Result

	Desc       string
	Locale     string
	Ammenities []string
	Refs       []Ref
}
