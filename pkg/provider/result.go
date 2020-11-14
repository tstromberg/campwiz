package provider

import (
	"time"
)

// Result is supposed to be a vendor neutral result of results
type Result struct {
	Name         string
	Distance     float64
	State        string
	ShortDesc    string
	Availability []time.Time
	URL          string
	Amenities    string

	Contract string
	ParkID   int
}
