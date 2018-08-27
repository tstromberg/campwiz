package query

import (
	"log"
	"sort"

	"github.com/tstromberg/campwiz/data"
	"github.com/tstromberg/campwiz/result"
)

// merge merges multiple dates together and multiple datasets.
func merge(res result.Results) result.Results {
	log.Printf("Merging %d results ...", len(res))
	m := make(map[string]result.Result)
	for _, r := range res {
		key := r.SiteKey()
		// Merge availability data.
		if val, exists := m[key]; exists {
			log.Printf("%s: Appending availability: %+v (previous: %+v)", key, r.Availability, val.Availability)
			val.Availability = append(val.Availability, r.Availability...)
			// map items are immutable.
			m[key] = val
			log.Printf("%s availability now: %+v", key, m[key].Availability)
		} else {
			log.Printf("%s: Not yet seen: %+v", key, r)
			data.Merge(&r)
			m[key] = r
		}
	}

	var merged result.Results
	for k, v := range m {
		log.Printf("%s: %+v", k, v)
		merged = append(merged, v)
	}
	sort.Sort(merged)
	return merged
}
