package metasrc

import (
	"bufio"
	"html"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/tstromberg/campwiz/pkg/campwiz"
	"github.com/tstromberg/campwiz/pkg/mangle"
	"k8s.io/klog/v2"
)

var (
	ccTitleRe  = regexp.MustCompile(`^<h3 class="h3_1".*<strong>(.*?)</strong></a></h3>`)
	ccRatingRe = regexp.MustCompile(`Scenic rating: (\d+)`)
	ccDescRe   = regexp.MustCompile(`<p class="noindent">(.*?)</p>`)
	ccLocaleRe = regexp.MustCompile(`^<p class="noindentt_1">(.*?)</p>`)
	tagRe      = regexp.MustCompile(`<.*?>`)
	//	ccFacilitiesRe = regexp.MustCompile(`<p class="noindent"><strong>Campsites, facilities:</strong>(.*?)</p>`)
	//	ccReserveRe    = regexp.MustCompile(`<p class="noindent"><strong>Reservations, fees:</strong> (.*?)</p>`)
	//	ccContactRe    = regexp.MustCompile(`^<p class="noindent"><strong>Contact:</strong> (.*?)</p>`)
)

// ccKey returns a "unique" string for a campground.
func ccKey(name string, locale string) string {
	key := name
	var shortened bool
	for {
		key, shortened = mangle.Shorten(key)
		if !shortened {
			break
		}
	}

	var location []string
	for _, word := range strings.Split(locale, " ") {
		if word == strings.Title(word) {
			if _, exists := mangle.ExtraWords[strings.ToUpper(word)]; !exists {
				location = append(location, word)
			}
		} else if len(location) > 1 {
			location = []string{}
		}
	}

	if len(location) > 2 {
		location = location[0:2]
	}

	return "/cc/" + strings.ToLower(location[0]) + "/" + strings.ToLower(strings.Replace(key, " ", "_", -1))
}

func htmlText(s string) string {
	return html.UnescapeString(tagRe.ReplaceAllString(s, ""))
}

// CC scans CC HTML, emits cross-references
func CC(r io.Reader) ([]campwiz.Ref, error) {
	scanner := bufio.NewScanner(r)
	seen := make(map[string]bool)

	var entries []campwiz.Ref
	var xref campwiz.Ref

	for scanner.Scan() {
		line := scanner.Text()
		klog.V(1).Infof("Line: %s", line)

		m := ccTitleRe.FindStringSubmatch(line)
		if len(m) > 0 {
			klog.V(1).Infof("Title: %s", m[1])

			// Clear the previous entry.
			if xref.Name != "" && xref.Rating > 0 {
				xref.ID = ccKey(xref.Name, xref.Locale)
				if _, exists := seen[xref.ID]; exists {
					klog.Warningf("Ignoring duplicate: %s (its ok)", xref.ID)
					continue
				}
				seen[xref.ID] = true
				entries = append(entries, xref)
			}

			xref = campwiz.Ref{Name: mangle.Title(htmlText(m[1]))}
			continue
		}

		m = ccRatingRe.FindStringSubmatch(line)
		if len(m) > 0 {
			klog.V(1).Infof("Rating: %s", m[1])
			var err error
			xref.Rating, err = strconv.ParseFloat(m[1], 64)
			if err != nil {
				klog.Errorf("unable to parse float %q: %v", m[1], err)
			}
			continue
		}
		m = ccLocaleRe.FindStringSubmatch(line)
		if xref.Rating > 0 && len(m) > 0 {
			klog.V(1).Infof("Locale: %s", m[1])
			xref.Locale = mangle.Locale(htmlText(m[1]))
		}

		m = ccDescRe.FindStringSubmatch(line)
		if xref.Rating > 0 && len(m) > 0 {
			klog.V(1).Infof("Desc: %s", m[1])
			if xref.Desc == "" {
				xref.Desc = htmlText(m[1])
			}
			continue
		}
	}

	// Close out the final parsed entry
	if xref.Name != "" {
		xref.ID = ccKey(xref.Name, xref.Locale)
		if _, exists := seen[xref.ID]; exists {
			klog.Warningf("Ignoring duplicate: %s (its ok)", xref.ID)
		} else {
			entries = append(entries, xref)
		}
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].ID < entries[j].ID })
	return entries, nil
}
