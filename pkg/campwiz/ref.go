package campwiz

// Ref is a cross-reference entry
type Ref struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name,omitempty"`

	Related []string `yaml:"related,omitempty"`
	URLs    []string `yaml:"urls,omitempty"`

	Rating   float64  `yaml:"rating,omitempty"`
	Tags     []string `yaml:"tags,omitempty"`
	Features []string `yaml:"features,omitempty"`
	Desc     string   `yaml:"desc,omitempty"`
	Owner    string   `yaml:"owner,omitempty"`

	Locale string  `yaml:"locale,omitempty"`
	Lat    float64 `yaml:"lat,omitempty"`
	Lon    float64 `yaml:"lon,omitempty"`

	// Synthetic part of the data
	Source RefSource `yaml:"__src,omitempty"`
}

type RefSource struct {
	Name       string
	RatingDesc string  `yaml:"rating_desc"`
	RatingMax  float64 `yaml:"rating_max"`
}

// RefData is cross-reference data loaded from YAML
type RefData struct {
	Source  RefSource
	Entries []Ref
}
