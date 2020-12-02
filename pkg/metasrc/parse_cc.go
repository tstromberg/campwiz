package metasrc

import (
	"bufio"
	"html"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/tstromberg/campwiz/pkg/campwiz"
	"github.com/tstromberg/campwiz/pkg/mangle"
	"k8s.io/klog/v2"
)

var (
	ccTitleRe      = regexp.MustCompile(`^<h3 class="h3_1".*<strong>(.*?)</strong></a></h3>`)
	ccRatingRe     = regexp.MustCompile(`Scenic rating: (\d+)`)
	ccDescRe       = regexp.MustCompile(`<p class="noindent">(.*?)</p>`)
	ccLocaleRe     = regexp.MustCompile(`^<p class="noindentt_1">(.*?)</p>`)
	tagRe          = regexp.MustCompile(`<.*?>`)
	ccFacilitiesRe = regexp.MustCompile(`<p class="noindent"><strong>Campsites, facilities:</strong>(.*?)</p>`)
	ccReserveRe    = regexp.MustCompile(`<p class="noindent"><strong>Reservations, fees:</strong> (.*?)</p>`)
	ccContactRe    = regexp.MustCompile(`^<p class="noindent"><strong>Contact:</strong> (.*?)</p>`)
	hrefRe         = regexp.MustCompile(`href="(.*?)">`)
	nonWordRe      = regexp.MustCompile(`\W+`)
)

// ccPropertyKey returns a "unique" string for a property.
func ccPropertyKey(name string, locale string) string {
	key := name
	var shortened bool
	for {
		key, shortened = mangle.Shorten(key)
		if !shortened {
			break
		}
	}

	var location []string

	short := mangle.ShortLocale(locale)
	for _, word := range strings.Split(short, " ") {
		if word == strings.Title(word) {
			if _, exists := mangle.ExtraWords[strings.ToUpper(word)]; !exists {
				location = append(location, word)
			}
		} else if len(location) > 1 {
			location = []string{}
		}
	}

	if len(location) > 4 {
		location = location[0:4]
	}

	newloc := strings.Join(location, "_")
	newloc = strings.Replace(newloc, "/", " ", -1)
	base := "/ca/" + strings.ToLower(nonWordRe.ReplaceAllString(newloc, ""))
	key = strings.ToLower(nonWordRe.ReplaceAllString(strings.Replace(key, " ", "_", -1), ""))

	// The property is the locale
	if mangle.LocaleProperty(locale) != "" {
		return base
	}

	return base + "/" + key
}

func campKey(name string, propertyKey string) string {
	key := strings.ToLower(strings.Replace(name, " ", "_", -1))
	if strings.HasSuffix(propertyKey, key) {
		return "default"
	}
	return key
}

func htmlText(s string) string {
	return strings.TrimSpace(html.UnescapeString(tagRe.ReplaceAllString(s, "")))
}

func finalizeProp(p *campwiz.Property, ref *campwiz.Ref) *campwiz.Property {
	klog.Infof("finalizing: prop: %+v ref: %+v", p, ref)
	p.Campgrounds[0].Refs = map[string]*campwiz.Ref{"cc": ref}
	fields := strings.Split(ref.Contact, ",")
	p.ManagedBy = strings.TrimSpace(fields[0])
	p.ID = ccPropertyKey(ref.Name, ref.Locale)
	p.Campgrounds[0].ID = campKey(ref.Name, p.ID)
	p.Campgrounds[0].Name = ref.Name
	propertyName := mangle.LocaleProperty(ref.Locale)
	if propertyName != "" {
		p.Name = propertyName
	}

	return p
}

// CC scans CC HTML, emits cross-references
func CC(r io.Reader, props map[string]*campwiz.Property) error {
	scanner := bufio.NewScanner(r)

	var prop *campwiz.Property
	var ref *campwiz.Ref

	for scanner.Scan() {
		line := scanner.Text()
		klog.V(1).Infof("Line: %s", line)

		m := ccTitleRe.FindStringSubmatch(line)
		if len(m) > 0 {
			klog.V(1).Infof("Title: %s", m[1])
			if prop != nil {
				final := finalizeProp(prop, ref)
				found, ok := props[final.ID]
				if !ok {
					props[final.ID] = final
					klog.Errorf("added %q: %v", final.ID, []byte(final.ID))
				} else {
					klog.Errorf("adding to %q / %s: %+v", final.ID, []byte(final.ID), final.Campgrounds)
					found.Campgrounds = append(found.Campgrounds, final.Campgrounds...)
				}
				prop = nil
				ref = nil
			}

			name := mangle.Title(htmlText(m[1]))
			ref = &campwiz.Ref{Name: name}
			prop = &campwiz.Property{Name: ref.Name, Campgrounds: []*campwiz.Campground{{ID: "default"}}}
			continue
		}

		// Nothing matters until we have a reference
		if ref == nil {
			continue
		}

		m = ccRatingRe.FindStringSubmatch(line)
		if m != nil {
			klog.V(1).Infof("Rating: %s", m[1])
			r, err := strconv.ParseFloat(m[1], 64)
			if err != nil {
				klog.Errorf("unable to parse float %q: %v", m[1], err)
			}
			ref.Rating = r
			continue
		}

		m = ccLocaleRe.FindStringSubmatch(line)
		if m != nil {
			klog.V(1).Infof("Locale: %s", m[1])
			ref.Locale = htmlText(m[1])
			continue
		}

		// Only match the first result
		if ref.Desc == "" {
			m = ccDescRe.FindStringSubmatch(line)
			if m != nil {
				klog.V(1).Infof("Desc: %s", m[1])
				ref.Desc = htmlText(m[1])
				continue
			}
		}

		m = ccReserveRe.FindStringSubmatch(line)
		if m != nil {
			klog.V(1).Infof("Reserve: %s", m[1])
			res := htmlText(m[1])
			if strings.Contains(res, "Reservations are not accepted") {
				klog.V(1).Infof("No reservations for %s", ref.Name)
				// Skip it
				ref = nil
				prop = nil
				continue
			}

			h := hrefRe.FindStringSubmatch(m[1])
			klog.V(1).Infof("href in %q: %v", m[1], h)
			if h != nil {
				klog.V(1).Infof("Found reserve URL: %s", h[1])
				prop.Campgrounds[0].ResURL = h[1]
			}
			continue
		}

		m = ccContactRe.FindStringSubmatch(line)
		if m != nil {
			klog.V(1).Infof("Contact: %s", m[1])
			ref.Contact = htmlText(m[1])

			h := hrefRe.FindStringSubmatch(m[1])
			klog.V(1).Infof("href in %q: %v", m[1], h)
			if h != nil {
				klog.V(1).Infof("Found contact URL: %s", m[1])
				prop.URL = h[1]
			}
			continue
		}

	}

	// Close out the final parsed entry
	if ref != nil {
		final := finalizeProp(prop, ref)
		found := props[final.ID]
		if found == nil {
			props[final.ID] = final
		} else {
			klog.Errorf("adding to %q: %v", final.ID, final.Campgrounds)
			found.Campgrounds = append(found.Campgrounds, final.Campgrounds...)
		}
	}

	return nil
}
