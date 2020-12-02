package search

import (
	"github.com/tstromberg/campwiz/pkg/campwiz"
	"k8s.io/klog"
)

// filter applies post-fetch filtering
func filter(q campwiz.Query, rs []campwiz.Result) []campwiz.Result {
	fs := []campwiz.Result{}

	for _, r := range rs {
		if q.MaxDistance > 0 && r.Distance > float64(q.MaxDistance) {
			klog.Infof("filtering %q -- too far (%.0f miles)", r.Name, r.Distance)
			continue
		}

		if q.MinRating > r.Rating {
			klog.Infof("filtering %q -- too low of a rating: %.1f", r.Name, r.Rating)
			continue
		}

		if len(q.Keywords) > 0 {
			/*
				fields := []string{r.Result.Desc, r.Result.Name}
				fields = append(fields, r.Result.Features...)
				for _, x := range r.Refs {
					fields = append(fields, x.Name, x.Locale, x.Desc, x.Owner)
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
					klog.Infof("filtering %q -- does not match %v", r.Name, q.Keywords)
					continue
				}
			*/
		}
		fs = append(fs, r)
	}
	return fs
}
