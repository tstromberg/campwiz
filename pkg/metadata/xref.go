package metadata

import "github.com/tstromberg/campwiz/pkg/campwiz"

var CompressHeader = `H4sIAAAAAAAA/`
var CompressPrefix = `z`

// Load returns all cross-reference data
func Load() (map[string]campwiz.Ref, error) {
	return LoadCC()
}
