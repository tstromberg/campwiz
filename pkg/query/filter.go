package query

import (
	"github.com/labstack/gommon/log"
	"github.com/tstromberg/campwiz/pkg/result"
	"k8s.io/klog/v2"
)

// filter applies post-fetch criteria filtering.
func filter(c Criteria, res result.Results) result.Results {
	klog.V(1).Infof("Filtering %d results ...", len(res))
	var filtered result.Results

	for _, r := range res {
		klog.Infof("Filtering %s: %+v against %+v", r.Name, r, c)
		if c.MaxDistance < int(r.Distance) {
			klog.Infof("%s is too far (%.0f miles)", r.Name, r.Distance)
			continue
		}
		if c.IncludeGroup && r.Availability[0].Group > 0 {
			filtered = append(filtered, r)
			log.Infof("Passes group filter: %+v", r.Name)
			continue
		}
		if c.IncludeBoatIn && r.Availability[0].Boat > 0 {
			filtered = append(filtered, r)
			log.Infof("Passes boat filter: %+v", r.Name)
			continue
		}
		if c.IncludeWalkIn && r.Availability[0].WalkIn > 0 {
			filtered = append(filtered, r)
			log.Infof("Passes walk-in filter: %+v", r.Name)
			continue
		}
		if c.IncludeStandard && r.Availability[0].Standard > 0 {
			filtered = append(filtered, r)
			log.Infof("Passes standard filter: %+v", r.Name)
			continue
		}
	}
	return filtered
}
