package campwiz

type Property struct {
	ID          string        `yaml:"id"` // Must be unique, suggested form: /<state>/<area>/<name>
	URL         string        `yaml:"url,omitempty"`
	Name        string        `yaml:"name"`
	ManagedBy   string        `yaml:"managed_by,omitempty"`
	Campgrounds []*Campground `yaml:"campgrounds"`
}

type Campground struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
	URL  string `yaml:"url,omitempty"`

	ResURL string `yaml:"res_url,omitempty"`
	ResID  string `yaml:"res_id,omitempty"`

	Refs map[string]*Ref

	PropertyID string `yaml:"__property_id__,omitempty"` // internal reference back
}

// Ref is a cross-reference entry
type Ref struct {
	URL string `yaml:"url,omitempty"`

	Name    string `yaml:"name,omitempty"`
	Desc    string `yaml:"desc,omitempty"`
	Contact string `yaml:"contact,omitempty"`

	Lat      float64  `yaml:"lat,omitempty"`
	Lon      float64  `yaml:"lon,omitempty"`
	Rating   float64  `yaml:"rating,omitempty"`
	Features []string `yaml:"features,omitempty"`
	Locale   string   `yaml:"locale,omitempty"`

	Lists []RefList `yaml:"lists,omitempty"`
}

// RefList is basically an award referenced by this site
type RefList struct {
	URL   string `yaml:"url,omitempty"`
	Title string `yaml:title,omitempty"`
	Place int    `yaml:place,omitempty"`
}

type Source struct {
	ID        string
	Name      string
	URL       string
	RatingMax float64 `yaml:"rating_max"`
}

type RefFile struct {
	Sources    map[string]Source `yaml:"sources,omitempty"`
	Properties []*Property
}
