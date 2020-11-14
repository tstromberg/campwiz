// utility to update cc.yaml
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/tstromberg/campwiz/pkg/mixer"
	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
)

var (
	titleRe      = regexp.MustCompile(`^\d+ ([A-Z].*)`)
	ratingRe     = regexp.MustCompile(`^Scenic rating: (\d+)`)
	descRe       = regexp.MustCompile(`^([A-Z].*[\.\!\)])$`)
	localeRe     = regexp.MustCompile(`^([a-z]{2,} .*[a-z])`)
	maxDescWords = 50
)

// key returns a "unique" string for a compground.
func key(name string, locale string) string {
	key := name
	var shortened bool
	for {
		key, shortened = mixer.ShortenName(key)
		if !shortened {
			break
		}
	}

	var location []string
	for _, word := range strings.Split(locale, " ") {
		if word == strings.Title(word) {
			if _, exists := mixer.ExtraWords[strings.ToUpper(word)]; !exists {
				location = append(location, word)
			}
		} else if len(location) > 1 {
			location = []string{}
		}
	}

	if len(location) > 2 {
		location = location[0:2]
	}
	return key + " - " + strings.Join(location, " ")
}

// parse parses text, emits entries.
func parse(scanner *bufio.Scanner) (mixer.CCData, error) {
	var e mixer.CCData
	seen := make(map[string]bool)

	s := mixer.XRef{}
	for scanner.Scan() {
		line := scanner.Text()
		klog.V(1).Infof("Line: %s", line)
		m := titleRe.FindStringSubmatch(line)
		if len(m) > 0 && line == strings.ToUpper(line) {
			klog.V(1).Infof("Title: %s", m[1])
			// Clear the previous entry.
			if s.Name != "" && s.Rating > 0 {
				s.Key = key(s.Name, s.Locale)
				if _, exists := seen[s.Key]; exists {
					klog.V(1).Infof("Ignoring duplicate: %s (its ok)", s.Key)
					continue
				}
				seen[s.Key] = true
				e.Entries = append(e.Entries, s)
			}
			s = mixer.XRef{Name: m[1]}
			continue
		}
		m = ratingRe.FindStringSubmatch(line)
		if len(m) > 0 {
			klog.V(1).Infof("Rating: %s", m[1])
			s.Rating, _ = strconv.Atoi(m[1])
			continue
		}
		m = localeRe.FindStringSubmatch(line)
		if s.Rating > 0 && len(m) > 0 {
			klog.V(1).Infof("Locale: %s", m[1])
			s.Locale = m[1]
		}

		m = descRe.FindStringSubmatch(line)
		if s.Rating > 0 && len(m) > 0 {
			klog.V(1).Infof("Desc: %s", m[1])
			if s.Desc == "" {
				s.Desc = m[1]
			} else if len(strings.Split(s.Desc, " ")) < maxDescWords {
				s.Desc = s.Desc + " " + m[1]
			}
			words := strings.Split(s.Desc, " ")
			if len(words) > maxDescWords {
				s.Desc = strings.Join(words[0:maxDescWords], " ") + " ..."
			}
			continue
		}

	}
	s.Key = key(s.Name, s.Locale)
	if _, exists := seen[s.Key]; exists {
		return e, fmt.Errorf("%s was already seen", s.Key)
	}
	seen[s.Key] = true
	e.Entries = append(e.Entries, s)
	return e, nil
}

func main() {
	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	es, err := parse(scanner)
	if err != nil {
		log.Fatalf("parse error: %v", err)
	}

	d, err := yaml.Marshal(&es)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("%s", d)

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
