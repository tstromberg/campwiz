package mixer

// XRef is a cross-reference entry
type XRef struct {
	SiteID string // NOTE: Currently unused

	Key    string // Key within the source
	Source string // Name of source

	Name   string // Name of site
	Rating int    // 0-9 rating
	Desc   string // Description
	Locale string // Long location info

	ShortLocale string // Shortened location information
}
