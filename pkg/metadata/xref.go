package metadata

// XRef is a cross-reference entry
type XRef struct {
	ID   string // `yaml:"id"`
	Name string // `yaml:"name,omitempty"`

	Related []string // `yaml:"related,omitempty"`
	URLs    []string // `yaml:"urls,omitempty"`

	Rating     float64  // `yaml:"rating,omitempty"`
	Tags       []string // `yaml:"tags,omitempty"`
	Facilities []string // `yaml:"facilities,omitempty"`
	Desc       string   // `yaml:"desc,omitempty"`
	Owner      string   // `yaml:"owner,omitempty"`

	Locale string  // `yaml:"locale,omitempty"`
	Lat    float64 // `yaml:"lat,omitempty"`
	Lon    float64 // `yaml:"lon,omitempty"`

	// Synthetic part of the data
	Source XrefSource // `yaml:"omitempty"`
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
