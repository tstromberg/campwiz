package provider

import (
	"fmt"

	"k8s.io/klog/v2"
)

// siteKey returns a unique key for a specific site.
func siteKey(r Result) string {
	return fmt.Sprintf("%s:%d:%s", r.Contract, r.ParkID, r.Name)
}

// merge merges multiple dates together
func mergeDates(res []Result) []Result {
	klog.V(1).Infof("Merging %d results ...", len(res))
	m := make(map[string]Result)
	for _, r := range res {
		key := siteKey(r)
		// Merge availability mixer.
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

// Search performs a RA, returns parsed results.
func Search(q Query) ([]Result, error) {
	var results []Result
	for _, d := range q.Dates {
		dr, err := searchRA(q, d)
		if err != nil {
			return results, err
		}
		results = append(results, dr...)
	}
	return mergeDates(results), nil
}
