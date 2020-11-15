package mixer

import (
	"io/ioutil"

	"github.com/tstromberg/campwiz/pkg/provider"
	"github.com/tstromberg/campwiz/pkg/relpath"
	"k8s.io/klog/v2"

	"gopkg.in/yaml.v2"
)

func findXRefs(r provider.Result, sources []XRef) []XRef {
	return nil
	/*
		variations := []string{
			r.Name,
			strings.Join(strings.Split(ShortName(ExpandknownAcronyms(r.Name)), " "), ""),
			ShortName(r.Name),
			ExpandknownAcronyms(r.Name),
			ShortName(ExpandknownAcronyms(r.Name)),
		}

		klog.V(2).Infof("Merge Variations: %v", strings.Join(variations, "|"))

		for _, name := range variations {
			mm := MMatches(name)
			klog.V(2).Infof("MMatches(%s) result: %v", name, mm)
			if len(mm) > 1 {
				// So, we have multiple matches. Perhaps the locale will help?
				klog.V(2).Infof("No unique for %s: %+v", name, mm)
				for _, m := range mm {
					// private knowledge
					if strings.Contains(r.ShortDesc, strings.Split(m, " - ")[1]) {
						klog.V(2).Infof("Lucky desc match: %s", m)
						r.M = M[m]
						return
					}
				}
			} else if len(mm) == 1 {
				klog.V(2).Infof("Match: %+v", mm)
				r.M = M[mm[0]]
				return
			}
		}
	*/
}

// MMatches finds the most likely key name for a campsite.
func MMatches(name string, xrefs []XRef) []string {
	return nil
	/*
		keyName := strings.ToUpper(name)
		klog.V(1).Infof("MMatches(%s) ...", keyName)

		// Three levels of matches.
		var exact []string
		var prefix []string
		var contains []string
		var allWords []string
		var someWords []string
		var singleWord []string

		keywords := strings.Split(keyName, " ")

		for k := range xrefs {
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
					allWords = append(allWords, k)
				} else if len(wordMatches) > 1 {
					klog.V(2).Infof("Partial match for %s: %s (matches=%v)", keyName, k, wordMatches)
					someWords = append(someWords, k)
				} else if len(wordMatches) == 1 {
					klog.V(3).Infof("Found single word match for %s: %s (matches=%v)", keyName, k, wordMatches)
					singleWord = append(singleWord, k)
				}
				continue
			}
			if i == 0 {
				if strings.HasPrefix(k, keyName+" - ") {
					exact = append(exact, k)
					klog.V(2).Infof("Found exact match for %s: %s", keyName, k)
					continue
				}
				klog.V(2).Infof("Found prefix match for %s: %s", keyName, k)
				prefix = append(prefix, k)
				continue
			} else if i > 0 {
				klog.V(2).Infof("Found substring match for %s: %s", keyName, k)
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
	*/
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
	xs = append(xs, ccd.Entries...)
	return xs, nil
}
