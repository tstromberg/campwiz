package search

import (
	"strings"

	"github.com/tstromberg/campwiz/pkg/campwiz"
	"k8s.io/klog"
)

// filter applies post-fetch filtering
func filter(q campwiz.Query, rs []campwiz.Result) []campwiz.Result {
	fs := []campwiz.Result{}

	for _, r := range rs {
		if q.MaxDistance > 0 && r.Distance > float64(q.MaxDistance) {
			klog.V(1).Infof("filtering %q -- too far (%.0f miles)", r.Name, r.Distance)
			continue
		}

		if q.MinRating > r.Rating {
			klog.V(1).Infof("filtering %q -- too low of a rating: %.1f", r.Name, r.Rating)
			continue
		}

		if len(q.Keywords) > 0 {
			fields := []string{r.Desc, r.Name}
			fields = append(fields, r.Features...)
			for _, x := range r.KnownCampground.Refs {
				fields = append(fields, x.Name, x.Locale, x.Desc)
			}
			found := false
			for _, f := range fields {
				for _, k := range q.Keywords {
					if strings.Contains(strings.ToLower(f), strings.ToLower(k)) {
						found = true
					}
				}
			}
			if !found {
				klog.V(1).Infof("filtering %q -- does not match %v", r.Name, q.Keywords)
				continue
			}
		}
		fs = append(fs, r)
	}
	return fs
}
