package mixer

// XRef is a cross-reference entry
type XRef struct {
	ID     string  // UNIQ key within the source
	Name   string  // Name of site
	Rating float64 // How does this rate?

	Tags []string // Is this place notable for anything, like scenery?

	Desc        string // Description
	Locale      string // Long location info
	ShortLocale string // Shortened location information

	// Hard-coded mapping to a site ID
	SiteID string // NOTE: Currently unused

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
