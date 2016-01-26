// Mix in data from different sources.
package data

import (
	"log"
	"strings"

	"github.com/tstromberg/autocamper/result"
)

var (
	M map[string]result.MEntry

	Acronyms = map[string]string{
		"MT.": "MOUNT",
		"SB":  "STATE BEACH",
		"SRA": "STATE RECREATION AREA",
		"SP":  "STATE PARK",
		"CP":  "COUNTY PARK",
		"NP":  "NATIONAL PARK",
	}

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

func ExpandAcronyms(s string) string {
	var words []string
	for _, w := range strings.Split(s, " ") {
		if val, exists := Acronyms[strings.ToUpper(w)]; exists {
			words = append(words, val)
		} else {
			words = append(words, w)
		}
	}
	expanded := strings.Join(words, " ")
	if expanded != s {
		log.Printf("Expanded %s to: %s", s, expanded)
	}
	return expanded
}

func ShortenName(s string) (string, bool) {
	log.Printf("Shorten: %s", s)
	keyWords := strings.Split(ExpandAcronyms(s), " ")
	for i, kw := range keyWords {
		if _, exists := ExtraWords[strings.ToUpper(kw)]; exists {
			log.Printf("Removing extra word in %s: %s", s, kw)
			keyWords = append(keyWords[:i], keyWords[i+1:]...)
			return strings.Join(keyWords, " "), true
		}
	}
	return s, false
}

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

func Merge(r *result.Result) {
	log.Printf("Merge: %s", r.Name)

	mm := MMatches(r)
	if len(mm) > 1 {
		// So, we have multiple matches. Perhaps the locale will help?
		log.Printf("No unique for %s: %+v", r.Name, mm)
		for _, m := range mm {
			// private knowledge
			if strings.Contains(r.ShortDesc, strings.Split(m, " - ")[1]) {
				log.Printf("Lucky desc match: %s", m)
				r.M = M[m]
				break
			}
		}
		return
	} else if len(mm) == 1 {
		log.Printf("Match: %+v", mm)
		r.M = M[mm[0]]
		return
	}
	log.Printf("Unable to match: %+v", r)
}
