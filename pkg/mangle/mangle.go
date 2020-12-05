// Package mangle performs text manipulation
package mangle

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
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

	// nonWords
	nonWordRe = regexp.MustCompile(`\W+`)
	// extra space
	spaceRe = regexp.MustCompile(`\s+`)
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
	return strings.Join(words, " ")
}

// Shorten is a one-pass shortening
func Shorten(s string) (string, bool) {
	keyWords := strings.Split(Expand(s), " ")
	for i, kw := range keyWords {
		if _, exists := ExtraWords[strings.ToUpper(kw)]; exists {
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

// Normalize removes weird characters so that a string is easy to compare
func Normalize(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	s, _, _ = transform.String(t, s)
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.Replace(s, `'`, "", -1)
	s = nonWordRe.ReplaceAllString(s, " ")
	return spaceRe.ReplaceAllLiteralString(s, " ")
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
