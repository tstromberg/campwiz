package mixer

import (
	"io/ioutil"
	"strings"

	"github.com/tstromberg/campwiz/pkg/provider"
	"github.com/tstromberg/campwiz/pkg/relpath"
	"k8s.io/klog/v2"

	"gopkg.in/yaml.v2"
)

func findXRefs(r provider.Result, xrefs []XRef) []XRef {
	var matching []XRef

	for _, xref := range xrefs {
		if xref.SiteID == r.ID {
			matching = append(matching, xref)
			continue
		}
	}

	variations := []string{
		r.Name,
		strings.Join(strings.Split(ShortName(expandAcronyms(r.Name)), " "), ""),
		ShortName(r.Name),
		expandAcronyms(r.Name),
		ShortName(expandAcronyms(r.Name)),
	}

	klog.V(2).Infof("Merge Variations: %v", strings.Join(variations, "|"))

	for _, name := range variations {
		mm := fuzzyMatch(name, xrefs)
		if len(mm) == 0 {
			continue
		}

		if len(mm) == 1 {
			return mm
		}

		if len(mm) > 1 {
			// So, we have multiple matches. Perhaps the locale will help? We no longer have it :(
			// BETTER IDEA: Fuzzy coordinates match?
			klog.V(2).Infof("No unique for %s: %+v - returning all", name, mm)
			return mm
		}
	}

	return matching
}

// fuzzyMatch finds the most likely matching cross-references for a site by name
func fuzzyMatch(name string, xrefs []XRef) []XRef {
	keyName := strings.ToUpper(name)
	klog.V(1).Infof("fuzzyMatch(%s) ...", keyName)

	// Three levels of matches.
	var exact []XRef
	var prefix []XRef
	var contains []XRef
	var allWords []XRef
	var someWords []XRef
	var singleWord []XRef

	keywords := strings.Split(keyName, " ")

	for _, xref := range xrefs {
		k := xref.ID
		i := strings.Index(k, keyName)
		klog.V(4).Infof("Testing: keyName=%s == k=%s (index=%d)", keyName, k, i)
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
				klog.V(2).Infof("All words match for %s: %s", keyName, k)
				allWords = append(allWords, xref)
			} else if len(wordMatches) > 1 {
				klog.V(2).Infof("Partial match for %s: %s (matches=%v)", keyName, k, wordMatches)
				someWords = append(someWords, xref)
			} else if len(wordMatches) == 1 {
				klog.V(3).Infof("Found single word match for %s: %s (matches=%v)", keyName, k, wordMatches)
				singleWord = append(singleWord, xref)
			}
			continue
		}
		if i == 0 {
			if strings.HasPrefix(k, keyName+" - ") {
				exact = append(exact, xref)
				klog.V(2).Infof("Found exact match for %s: %s", keyName, k)
				continue
			}
			klog.V(2).Infof("Found prefix match for %s: %s", keyName, k)
			prefix = append(prefix, xref)
			continue
		} else if i > 0 {
			klog.V(2).Infof("Found substring match for %s: %s", keyName, k)
			contains = append(contains, xref)
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

// LoadCC returns CC cross-reference data
func LoadCC() ([]XRef, error) {
	p := relpath.Find("data/cc.yaml")
	f, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}

	var ccd XrefData
	err = yaml.Unmarshal(f, &ccd)
	if err != nil {
		return nil, err
	}

	klog.V(1).Infof("Loaded %d entries from %s ...", len(ccd.Entries), p)

	var xs []XRef
	for _, e := range ccd.Entries {
		e.Source = ccd.Source
		xs = append(xs, e)
	}
	return xs, nil
}
