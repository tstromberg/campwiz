package search

import (
	"fmt"
	"sort"
	"strings"

	"github.com/tstromberg/campwiz/pkg/campwiz"
	"github.com/tstromberg/campwiz/pkg/mangle"
	"k8s.io/klog"
)

const (
	NoMatch = iota
	DoubleMangledSubMatch
	MangledSubMatch
	SubNameMatch
	DoubleMangledNameMatch
	MangledNameMatch
	NameMatch
	SiteID
)

var (
	scoreNames = map[int]string{
		NoMatch:                "NoMatch",
		DoubleMangledSubMatch:  "DoubleMangledSubMatch",
		MangledSubMatch:        "MangledSubMatch",
		SubNameMatch:           "SubNameMatch",
		DoubleMangledNameMatch: "DoubleMangledNameMatch",
		MangledNameMatch:       "MangledNameMatch",
		NameMatch:              "NameMatch",
		SiteID:                 "SiteID",
	}
)

type Match struct {
	Score      int
	Detail     string
	Campground *campwiz.Campground
}

func annotate(r campwiz.Result, props map[string]*campwiz.Property) campwiz.Result {
	cg := findBestMatch(r, props)
	if cg.Score == 0 {
		klog.Errorf("No site matche for %+v", r)
		return r
	}
	r.KnownCampground = cg.Campground

	for _, ref := range cg.Campground.Refs {
		if r.Locale == "" && ref.Locale != "" {
			r.Locale = ref.Locale
		}
	}
	return r
}

func findBestMatch(r campwiz.Result, props map[string]*campwiz.Property) Match {
	matches := findMatches(r, props)

	if len(matches) == 0 {
		return Match{Score: NoMatch}
	}

	sort.Slice(matches, func(i, j int) bool { return matches[i].Score > matches[j].Score })
	return matches[0]
}

func findMatches(r campwiz.Result, props map[string]*campwiz.Property) []Match {
	var matches []Match
	resName := strings.TrimSpace(strings.ToLower(r.Name))

	for _, prop := range props {
		for _, cg := range prop.Campgrounds {
			knownName := strings.ToLower(cg.Name)

			if resName == knownName {
				matches = append(matches, Match{NameMatch, fmt.Sprintf("result %q = known %q", knownName, resName), cg})
				continue
			}

			if strings.Contains(resName, knownName) {
				matches = append(matches, Match{SubNameMatch, fmt.Sprintf("known %q in result %q", knownName, resName), cg})
			}

			if strings.Contains(knownName, resName) {
				matches = append(matches, Match{SubNameMatch, fmt.Sprintf("result %q in known %q", resName, knownName), cg})
			}

			// Mangle the result to see if it roughly matches a known site
			resVariations := []string{
				strings.Join(strings.Split(mangle.Shortest(mangle.Expand(resName)), " "), ""),
				mangle.Shortest(resName),
				mangle.Expand(resName),
				mangle.Shortest(mangle.Expand(resName)),
			}

			// Mangle the result to see if it roughly matches a known site
			knownVariations := []string{
				strings.Join(strings.Split(mangle.Shortest(mangle.Expand(knownName)), " "), ""),
				mangle.Shortest(knownName),
				mangle.Expand(knownName),
				mangle.Shortest(mangle.Expand(knownName)),
			}

			for i, rv := range resVariations {
				rv = strings.ToLower(rv)
				klog.Errorf("rv %q vs known %q", rv, knownName)
				if rv == knownName {
					matches = append(matches, Match{MangledNameMatch, fmt.Sprintf("variation %d: %q = %q", i, rv, knownName), cg})
					continue
				}
				if strings.Contains(knownName, rv) {
					matches = append(matches, Match{MangledSubMatch, fmt.Sprintf("variation %d: result %q in known %q", i, rv, knownName), cg})
					continue
				}
				if strings.Contains(rv, knownName) {
					matches = append(matches, Match{MangledSubMatch, fmt.Sprintf("variation %d: result %q in known %q", i, knownName, rv), cg})
					continue
				}

				for x, kv := range knownVariations {
					kv = strings.ToLower(kv)
					klog.Errorf("rv %q vs kv %q", rv, kv)
					if rv == kv {
						matches = append(matches, Match{DoubleMangledNameMatch, fmt.Sprintf("variation %d/%d: %q = %q", i, x, rv, knownName), cg})
						continue
					}
					if strings.Contains(kv, rv) {
						matches = append(matches, Match{DoubleMangledSubMatch, fmt.Sprintf("variation %d/%d: result %q in known %q", i, x, rv, kv), cg})
						continue
					}
					if strings.Contains(rv, kv) {
						matches = append(matches, Match{DoubleMangledSubMatch, fmt.Sprintf("variation %d/%d: result %q in known %q", i, x, kv, rv), cg})
						continue
					}
				}
			}
		}
	}

	return matches
}
