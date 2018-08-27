package query

import (
	"github.com/golang/glog"
	"github.com/labstack/gommon/log"
	"github.com/tstromberg/campwiz/result"
)

// filter applies post-fetch criteria filtering.
func filter(c Criteria, res result.Results) result.Results {
	glog.V(1).Infof("Filtering %d results ...", len(res))
	var filtered result.Results

	for _, r := range res {
		glog.Infof("Filtering %s: %+v against %+v", r.Name, r, c)
		if c.MaxDistance < int(r.Distance) {
			glog.Infof("%s is too far (%.0f miles)", r.Name, r.Distance)
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
