package search

import (
	"time"

	"github.com/tstromberg/campwiz/pkg/cache"
	"github.com/tstromberg/campwiz/pkg/metadata"
	"k8s.io/klog/v2"
)

var (
	// searchPageExpiry is how long search pages can be cached for.
	searchPageExpiry = time.Duration(6*3600) * time.Second

	// amount of time to sleep between uncached fetches
	uncachedDelay = time.Millisecond * 600

	// maximum number of pages to fetch
	maxPages = 15
)

// siteKey returns a unique key for a specific site.
func siteKey(r Result) string {
	return r.ID
}

// merge merges multiple dates together
func mergeDates(res []Result) []Result {
	klog.V(1).Infof("Merging %d results ...", len(res))
	m := make(map[string]Result)
	for _, r := range res {
		key := siteKey(r)
		// Merge availability metadata.
		if val, exists := m[key]; exists {
			klog.V(1).Infof("%s: Appending availability: %+v (previous: %+v)", key, r.Availability, val.Availability)
			val.Availability = append(val.Availability, r.Availability...)
			// map items are immutable.
			m[key] = val
			klog.V(1).Infof("%s availability now: %+v", key, m[key].Availability)
		} else {
			klog.V(1).Infof("%s: Not yet seen: %+v", key, r)
			m[key] = r
		}
	}

	var merged []Result
	for k, v := range m {
		klog.V(1).Infof("%s: %+v", k, v)
		merged = append(merged, v)
	}
	return merged
}

// All performs a RA, returns parsed results.
func All(q Query, cs cache.Store, xrefs map[string]metadata.XRef) ([]Result, error) {
	klog.Infof("search query: %+v", q)

	var results []Result
	for _, d := range q.Dates {
		// TODO: Parallel search between providers
		dr, err := searchSCC(q, d, cs)
		if err != nil {
			klog.Errorf("searchSCC failed: %v", err)
		}
		results = append(results, dr...)

		dr, err = searchRC(q, d, cs)
		if err != nil {
			klog.Errorf("searchRC failed: %v", err)
		}
		results = append(results, dr...)

		dr, err = searchRA(q, d, cs)
		if err != nil {
			klog.Errorf("searchRA failed: %v", err)
		}
		results = append(results, dr...)

		dr, err = searchSMC(q, d, cs)
		if err != nil {
			klog.Errorf("searchSMC failed: %v", err)
		}
		results = append(results, dr...)
	}
	return mergeDates(results), nil
}
