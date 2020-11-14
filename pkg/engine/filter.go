package query

import (
	"k8s.io/klog/v2"
)

// filter applies post-fetch Query filtering.
func filter(c Query, res []engine.Result) []engine.Result {
	klog.V(1).Infof("Filtering %d results ...", len(res))
	var filtered []engine.Result

	for _, r := range res {
		klog.Infof("Filtering %s: %+v against %+v", r.Name, r, c)
		if c.MaxDistance < int(r.Distance) {
			klog.Infof("%s is too far (%.0f miles)", r.Name, r.Distance)
			continue
		}
		if c.IncludeGroup && r.Availability[0].Group > 0 {
			filtered = append(filtered, r)
			klog.Infof("Passes group filter: %+v", r.Name)
			continue
		}
		if c.IncludeBoatIn && r.Availability[0].Boat > 0 {
			filtered = append(filtered, r)
			klog.Infof("Passes boat filter: %+v", r.Name)
			continue
		}
		if c.IncludeWalkIn && r.Availability[0].WalkIn > 0 {
			filtered = append(filtered, r)
			klog.Infof("Passes walk-in filter: %+v", r.Name)
			continue
		}
		if c.IncludeStandard && r.Availability[0].Standard > 0 {
			filtered = append(filtered, r)
			klog.Infof("Passes standard filter: %+v", r.Name)
			continue
		}
	}
	return filtered
}
