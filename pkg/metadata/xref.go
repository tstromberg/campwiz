package metadata

// XRef is a cross-reference entry
type XRef struct {
	ID         string   // unique key across all sources
	SiteIDs    []string // Campsites this cross-reference applies to
	Name       string   // Name of site
	Rating     float64  // How does this rate?
	Tags       []string // Is this place notable for anything, like scenery?
	Ammenities []string
	Desc       string // Description

	Owner  string // name of owner
	Locale string // free-form location information

	Park   string // The park this is in, if any
	City   string // City the park is in
	County string // County the park is in

	Lat float64 // Latitude
	Lon float64 // Longitude

	// Synthetic part of the data
	Source XrefSource
}

type XrefSource struct {
	Name       string
	RatingDesc string  `yaml:"rating_desc"`
	RatingMax  float64 `yaml:"rating_max"`
}

// XrefData is cross-reference data loaded from YAML
type XrefData struct {
	Source  XrefSource
	Entries []XRef
}

// Load returns all cross-reference data
func Load() (map[string]XRef, error) {
	return LoadCC()
}
