package search

import (
	"strings"

	"github.com/tstromberg/campwiz/pkg/campwiz"
	"k8s.io/klog"
)

// filter applies post-fetch filtering
func filter(q campwiz.Query, as []campwiz.AnnotatedResult) []campwiz.AnnotatedResult {
	fs := []campwiz.AnnotatedResult{}

	for _, r := range as {
		if q.MaxDistance > 0 && r.Distance > float64(q.MaxDistance) {
			klog.Infof("filtering %q -- too far (%d miles)", r.Name, r.Distance)
			continue
		}

		if q.MinSceneryRating > 0 {
			sr := 0.0
			for _, x := range r.Refs {
				if strings.Contains(strings.ToLower(x.Source.RatingDesc), "scenery") {
					sr = x.Rating
				}
			}

			if sr > 0 && float64(q.MinSceneryRating) > sr {
				klog.Infof("filtering %q -- too ugly (%d scenery)", r.Name, sr)
				continue
			}
		}

		if len(q.Keywords) > 0 {
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
		}
		fs = append(fs, r)
	}
	return fs
}
