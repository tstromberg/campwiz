// M specific code.
package data

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/tstromberg/autocamper/result"

	"gopkg.in/yaml.v2"
)

type MEntries struct {
	Entries []result.MEntry
}

// MMatches finds the most likely key name for a campsite.
func MMatches(name string) []string {
	keyName := strings.ToUpper(name)
	log.Printf("MMatches(%s) ...", keyName)

	// Three levels of matches.
	var exact []string
	var prefix []string
	var contains []string
	var allWords []string
	var someWords []string
	var singleWord []string

	keywords := strings.Split(keyName, " ")

	for k, _ := range M {
		i := strings.Index(k, keyName)
		log.Printf("Testing: keyName=%s == k=%s (index=%d)", keyName, k, i)
		// The whole key does not exist.
		if i == -1 {
			var wordMatches []string
			kwords := strings.Split(k, " ")
			for _, kw := range kwords {
				for _, keyword := range keywords {
					if keyword == kw {
						wordMatches = append(wordMatches, kw)
					}
				}
			}
			if len(wordMatches) == len(keywords) {
				log.Printf("All words match for %s: %s", keyName, k)
				allWords = append(allWords, k)
			} else if len(wordMatches) > 1 {
				log.Printf("Partial match for %s: %s (matches=%v)", keyName, k, wordMatches)
				someWords = append(someWords, k)
			} else if len(wordMatches) == 1 {
				log.Printf("Found single word match for %s: %s (matches=%v)", keyName, k, wordMatches)
				singleWord = append(singleWord, k)
			}
			continue
		}
		if i == 0 {
			if strings.HasPrefix(k, keyName+" - ") {
				exact = append(exact, k)
				log.Printf("Found exact match for %s: %s", keyName, k)
				continue
			}
			log.Printf("Found prefix match for %s: %s", keyName, k)
			prefix = append(prefix, k)
			continue
		} else if i > 0 {
			log.Printf("Found substring match for %s: %s", keyName, k)
			contains = append(contains, k)
		}
	}

	if len(exact) > 0 {
		return exact
	}
	if len(prefix) > 0 {
		return prefix
	}
	if len(contains) > 0 {
		return contains
	}
	if len(allWords) > 0 {
		return allWords
	}
	if len(someWords) > 0 {
		return someWords
	}
	return singleWord
}

func LoadM(path string) error {
	M = make(map[string]result.MEntry)
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	var ms MEntries
	err = yaml.Unmarshal(f, &ms)
	if err != nil {
		return err
	}
	log.Printf("Loaded %d entries from %s ...", len(ms.Entries), path)
	for _, m := range ms.Entries {
		if val, ok := M[m.Key]; ok {
			return fmt.Errorf("%s already loaded. Previous=%+v, New=%+v", val, m)
		}
		M[m.Key] = m
		log.Printf("Loaded [%s]: %+v", m.Name, m)
	}
	return nil
}
