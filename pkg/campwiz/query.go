package campwiz

import "time"

// Query defines a list of attributes that can be sent to the camp engines
type Query struct {
	Lat         float64
	Lon         float64
	Dates       []time.Time
	StayLength  int
	MaxDistance int
	MinRating   float64
	Keywords    []string

	SiteKinds []int
	Features  []int
}
