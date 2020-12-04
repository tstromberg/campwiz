// Package mangle performs text manipulation
package mangle

import (
	"strings"

	"k8s.io/klog/v2"
)

var (
	knownAcronyms = map[string]string{
		"MT.": "MOUNT",
		"MT":  "MOUNT",
		"SB":  "STATE BEACH",
		"SRA": "STATE RECREATION AREA",
		"SP":  "STATE PARK",
		"CP":  "COUNTY PARK",
		"NP":  "NATIONAL PARK",
		"NF":  "NATIONAL FOREST",
		"SHP": "STATE HISTORIC PARK",
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
		"GROUP":      true,
		"VILLAGE":    true,
		"WALK-IN":    true,
		"RESORT":     true,
		"VISITOR":    true,
		"LONG":       true,
		"TERM":       true,
		"RV":         true,
		"SITES":      true,
		"FAMILY":     true,
		"SPA":        true,
		"ON":         true,
		"THE":        true,
		"BAY":        true,
		"BY":         true,
	}
)

func Expand(s string) string {
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

// Shorten is a one-pass shortening
func Shorten(s string) (string, bool) {
	klog.V(3).Infof("Shorten: %s", s)
	keyWords := strings.Split(Expand(s), " ")
	for i, kw := range keyWords {
		if _, exists := ExtraWords[strings.ToUpper(kw)]; exists {
			klog.V(1).Infof("Removing extra word in %s: %s", s, kw)
			keyWords = append(keyWords[:i], keyWords[i+1:]...)
			return strings.Join(keyWords, " "), true
		}
	}
	return s, false
}

// Shortest returns the shortest possible name for a string
func Shortest(s string) string {
	var shortened bool
	for {
		s, shortened = Shorten(s)
		if !shortened {
			break
		}
	}
	return s
}

// Locale returns a shorter locale name
func Locale(s string) string {
	new := []string{}

	for i, w := range strings.Split(s, " ") {
		lw := strings.ToLower(w)

		if i == 0 {
			switch lw {
			case "on", "in", "near":
				continue
			}
		}

		if i == 1 && lw == "the" {
			continue
		}

		if lw == "in" {
			new = append(new, "/")
			continue
		}

		new = append(new, w)
	}

	return strings.Join(new, " ")
}

func LocaleProperty(s string) string {

	words := strings.Split(s, " ")
	for i, w := range words {
		if w == "near" {
			words = words[0:i]
		}
	}

	for i, w := range words {
		switch w {
		case "/", "in":
			if words[i+1] == "the" {
				return strings.Join(words[i+2:], " ")
			}
			return strings.Join(words[i+1:], " ")
		}
	}
	return ""
}

// Locale returns an even shorter locale
func ShortLocale(s string) string {
	s = Locale(s)

	in := strings.Index(s, " / ")
	if in > 0 {
		s = s[in+3:]
	}

	new := []string{}
	for _, w := range strings.Split(s, " ") {
		lw := strings.ToLower(w)
		switch lw {
		case "national", "forest", "park", "recreation", "area", "state", "east", "of", "west", "north", "south", "at", "to", "demonstration", "off":
			continue
		}
		new = append(new, w)
	}
	return strings.Join(new, " ")
}

func Title(s string) string {
	new := []string{}
	for _, w := range strings.Split(strings.ToLower(s), " ") {
		switch w {
		case "a", "on", "in", "an", "the", "to", "at":
			new = append(new, w)
			continue
		case "rv", "svra", "nps":
			new = append(new, strings.ToTitle(w))
		default:
			new = append(new, strings.Title(w))
		}
	}
	return strings.Join(new, " ")
}

// Ellipsis sets a cap on the number of words to show
func Ellipsis(s string, max int) string {
	words := strings.Split(s, " ")
	if len(words) < max {
		return s
	}
	return strings.Join(words[0:max], " ") + " ..."
}
