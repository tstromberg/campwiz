package search

import (
	"fmt"
	"sort"
	"strings"

	"github.com/tstromberg/campwiz/pkg/campwiz"
	"github.com/tstromberg/campwiz/pkg/mangle"
	"github.com/tstromberg/campwiz/pkg/metadata"

	"github.com/agnivade/levenshtein"
	"k8s.io/klog/v2"
)

const (
	NoMatch = iota
	BiMangledPropSubMatch
	BiMangledPropMatch
	MangledPropSubMatch
	ApproxPropMatch
	MangledPropMatch
	PropMatch
	BiMangledSubMatch
	MangledSubMatch
	SubMatch
	SinglePropMatch
	ApproxMatch
	BiMangledMatch
	MangledMatch
	NameMatch
	SiteID
)

var scoreNames = map[int]string{
	NoMatch:               "NoMatch",
	BiMangledPropSubMatch: "BiMangledPropSubMatch",
	BiMangledPropMatch:    "BiMangledPropMatch",
	MangledPropSubMatch:   "MangledPropSubMatch",
	ApproxPropMatch:       "ApproxPropMatch",
	MangledPropMatch:      "MangledPropMatch",
	PropMatch:             "PropMatch",
	BiMangledSubMatch:     "BiMangledSubMatch",
	MangledSubMatch:       "MangledSubMatch",
	SubMatch:              "SubMatch",
	SinglePropMatch:       "SinglePropMatch",
	BiMangledMatch:        "BiMangledMatch",
	MangledMatch:          "MangledMatch",
	ApproxMatch:           "ApproxMatch",
	NameMatch:             "NameMatch",
	SiteID:                "SiteID",
}

type Match struct {
	Score      int
	Detail     string
	Campground *campwiz.Campground
}

func average(xs []float64) float64 {
	total := 0.0
	for _, v := range xs {
		total += v
	}
	return total / float64(len(xs))
}

func annotate(r campwiz.Result, props map[string]*campwiz.Property) campwiz.Result {
	cg := findBestMatch(r, props)
	if cg.Score == 0 {
		klog.Warningf("No site match for %+v", r)
		return r
	}
	r.KnownCampground = cg.Campground

	ratings := []float64{}

	for _, ref := range cg.Campground.Refs {
		if ref.Rating > 0 {
			// TODO: Take into account max
			ratings = append(ratings, ref.Rating)
		}
		if r.Locale == "" && ref.Locale != "" {
			r.Locale = ref.Locale
		}
		if r.Desc == "" && ref.Desc != "" {
			if ref.URL != "" {
				r.Desc = ref.Desc
			} else {
				r.Desc = mangle.Ellipsis(metadata.Decompress(ref.Desc), 65)
			}
		}
	}

	r.Rating = average(ratings)
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

var (
	varCache = map[string][]string{}
)

func variations(s string) []string {
	if varCache[s] != nil {
		return varCache[s]
	}
	try := map[string]bool{
		strings.ToLower(strings.Join(strings.Split(mangle.Shortest(mangle.Expand(s)), " "), "")): true,
		strings.ToLower(mangle.Shortest(s)):                true,
		strings.ToLower(mangle.Expand(s)):                  true,
		strings.ToLower(mangle.Shortest(mangle.Expand(s))): true,
	}

	vs := []string{}
	for k, _ := range try {
		vs = append(vs, k)
	}

	klog.Infof("variations for %q: %v", s, vs)
	varCache[s] = vs
	return varCache[s]
}

func findMatches(r campwiz.Result, props map[string]*campwiz.Property) []Match {
	var matches []Match
	resName := mangle.Normalize(r.Name)

	for _, prop := range props {
		propName := mangle.Normalize(prop.Name)
		// TODO: A better job guessing which campsite to use
		var cg *campwiz.Campground
		for _, c := range prop.Campgrounds {
			cg = c
		}

		if resName == propName {
			if len(prop.Campgrounds) == 1 {
				matches = append(matches, Match{SinglePropMatch, fmt.Sprintf("result %q = single park %q", resName, prop.Name), cg})
			} else {
				matches = append(matches, Match{PropMatch, fmt.Sprintf("result %q = multi park %q", resName, prop.Name), cg})
			}
		}

		for x, kv := range variations(propName) {
			if kv == resName {
				matches = append(matches, Match{MangledPropMatch, fmt.Sprintf("variation %d: %q = %q", x, kv, resName), cg})
			}

			for i, rv := range variations(resName) {
				if rv == kv {
					matches = append(matches, Match{BiMangledPropMatch, fmt.Sprintf("variation %d/%d: %q = %q", i, x, rv, propName), cg})
				}

				if strings.Contains(kv, rv) {
					matches = append(matches, Match{BiMangledPropSubMatch, fmt.Sprintf("variation %d/%d: result %q in known %q", i, x, rv, kv), cg})
					continue
				}
				if strings.Contains(rv, kv) {
					matches = append(matches, Match{BiMangledPropSubMatch, fmt.Sprintf("variation %d/%d: result %q in known %q", i, x, kv, rv), cg})
					continue
				}

				d := levenshtein.ComputeDistance(rv, kv)
				if d < 3 {
					matches = append(matches, Match{ApproxPropMatch, fmt.Sprintf("variation %d/%d: %q is %d edits from %q", i, x, rv, d, kv), cg})
					continue
				}
			}

		}

		for _, cg := range prop.Campgrounds {
			knownName := mangle.Normalize(cg.Name)

			if resName == knownName {
				matches = append(matches, Match{NameMatch, fmt.Sprintf("result %q = known %q", resName, knownName), cg})
				continue
			}

			if strings.Contains(resName, knownName) {
				matches = append(matches, Match{SubMatch, fmt.Sprintf("known %q in result %q", knownName, resName), cg})
			}

			if strings.Contains(knownName, resName) {
				matches = append(matches, Match{SubMatch, fmt.Sprintf("result %q in known %q", resName, knownName), cg})
			}

			for i, rv := range variations(resName) {
				if rv == knownName {
					matches = append(matches, Match{MangledMatch, fmt.Sprintf("variation %d: %q = %q", i, rv, knownName), cg})
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

				for x, kv := range variations(knownName) {
					kv = strings.ToLower(kv)
					if rv == kv {
						matches = append(matches, Match{BiMangledMatch, fmt.Sprintf("variation %d/%d: %q = %q", i, x, rv, knownName), cg})
						continue
					}
					if strings.Contains(kv, rv) {
						matches = append(matches, Match{BiMangledSubMatch, fmt.Sprintf("variation %d/%d: result %q in known %q", i, x, rv, kv), cg})
						continue
					}
					if strings.Contains(rv, kv) {
						matches = append(matches, Match{BiMangledSubMatch, fmt.Sprintf("variation %d/%d: result %q in known %q", i, x, kv, rv), cg})
						continue
					}

					d := levenshtein.ComputeDistance(rv, kv)
					if d < 3 {
						matches = append(matches, Match{ApproxMatch, fmt.Sprintf("variation %d/%d: %q is %d edits from %q", i, x, rv, d, kv), cg})
						continue
					}

				}
			}
		}
	}

	return matches
}
