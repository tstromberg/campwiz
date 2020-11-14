// Package query queries for campsites across providers
package query

import (
	"time"

	"github.com/tstromberg/campwiz/pkg/result"
	"k8s.io/klog/v2"
)

var (
	// searchPageExpiry is how long search pages can be cached for.
	searchPageExpiry = time.Duration(6*3600) * time.Second

	// the date format used
	campingDateFormat = "Mon Jan 2 2006"

	// amount of time to sleep between uncached fetches
	uncachedDelay = time.Millisecond * 750
)

// Criteria defines a list of attributes that can be sent to the camp engines
type Criteria struct {
	Lat         float64
	Lon         float64
	Dates       []time.Time
	Nights      int
	MaxDistance int
	MaxPages    int

	IncludeStandard bool
	IncludeGroup    bool
	IncludeBoatIn   bool
	IncludeWalkIn   bool
}

// Search performs a RA, returns parsed results.
func Search(crit Criteria) (result.Results, error) {
	var results result.Results
	for _, d := range crit.Dates {
		dr, err := searchRA(crit, d)
		if err != nil {
			return results, err
		}
		results = append(results, dr...)
	}
	klog.Infof("Found %d results", len(results))
	filtered := filter(crit, results)
	klog.Infof("Post-filter: %d results", len(filtered))
	merged := merge(filtered)
	klog.Infof("Post-merge: %d results", len(merged))
	return merged, nil
}
