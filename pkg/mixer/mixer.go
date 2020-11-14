package mixer

import (
	"strings"

	"github.com/tstromberg/campwiz/pkg/provider"
	"k8s.io/klog/v2"
)

var (
	knownAcronyms = map[string]string{
		"MT.": "MOUNT",
		"SB":  "STATE BEACH",
		"SRA": "STATE RECREATION AREA",
		"SP":  "STATE PARK",
		"CP":  "COUNTY PARK",
		"NP":  "NATIONAL PARK",
	}

	// ExtraWords are common words that can be thrown out for matching
	ExtraWords = map[string]bool{
		"&":          true,
		"(CA)":       true,
		"AND":        true,
		"AREA":       true,
		"CAMP":       true,
		"CAMPGROUND": true,
		"COUNTY":     true,
		"DAY":        true,
		"FOREST":     true,
		"FS":         true,
		"MONUMENT":   true,
		"NATIONAL":   true,
		"NATL":       true,
		"PARK":       true,
		"RECREATION": true,
		"REGIONAL":   true,
		"STATE":      true,
		"USE":        true,
	}
)

// MixedResult is a result with associated cross-reference data
type MixedResult struct {
	Result     provider.Result
	References []XRef
}

func expandAcronyms(s string) string {
	var words []string
	for _, w := range strings.Split(s, " ") {
		if val, exists := knownAcronyms[strings.ToUpper(w)]; exists {
			words = append(words, val)
		} else {
			words = append(words, w)
		}
	}
	expanded := strings.Join(words, " ")
	if expanded != s {
		klog.V(1).Infof("Expanded %s to: %s", s, expanded)
	}
	return expanded
}

// ShortenName is a one-pass shortening
func ShortenName(s string) (string, bool) {
	klog.V(3).Infof("Shorten: %s", s)
	keyWords := strings.Split(expandAcronyms(s), " ")
	for i, kw := range keyWords {
		if _, exists := ExtraWords[strings.ToUpper(kw)]; exists {
			klog.V(1).Infof("Removing extra word in %s: %s", s, kw)
			keyWords = append(keyWords[:i], keyWords[i+1:]...)
			return strings.Join(keyWords, " "), true
		}
	}
	return s, false
}

// ShortName returns the shortest possible name for a string
func ShortName(s string) string {
	var shortened bool
	for {
		s, shortened = ShortenName(s)
		if !shortened {
			break
		}
	}
	return s
}

// Mix combines results with cross-references
func Mix(results []provider.Result, xrefs []XRef) []MixedResult {
	ms := []MixedResult{}
	for _, r := range results {
		ms = append(ms, MixedResult{Result: r})
	}
	return ms
}
