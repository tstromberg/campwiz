// utility to update m.yaml
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/tstromberg/campwiz/data"
	"github.com/tstromberg/campwiz/result"
	"gopkg.in/yaml.v2"
)

var (
	titleRe      = regexp.MustCompile(`^\d+ ([A-Z].*)`)
	sRatingRe    = regexp.MustCompile(`^Scenic rating: (\d+)`)
	DescRe       = regexp.MustCompile(`^([A-Z].*[\.\!\)])$`)
	LocaleRe     = regexp.MustCompile(`^([a-z]{2,} .*[a-z])`)
	MaxDescWords = 50
)

// key returns a "unique" string for a compground.
func key(name string, locale string) string {
	key := name
	var shortened bool
	for {
		key, shortened = data.ShortenName(key)
		if !shortened {
			break
		}
	}

	var location []string
	for _, word := range strings.Split(locale, " ") {
		if word == strings.Title(word) {
			if _, exists := data.ExtraWords[strings.ToUpper(word)]; !exists {
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
func parse(scanner *bufio.Scanner) (data.MEntries, error) {
	var e data.MEntries
	seen := make(map[string]bool)

	s := result.MEntry{}
	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("Line: %s", line)
		m := titleRe.FindStringSubmatch(line)
		if len(m) > 0 && line == strings.ToUpper(line) {
			log.Printf("Title: %s", m[1])
			// Clear the previous entry.
			if s.Name != "" && s.SRating > 0 {
				s.Key = key(s.Name, s.Locale)
				if _, exists := seen[s.Key]; exists {
					log.Printf("Ignoring duplicate: %s (its ok)", s.Key)
					continue
				}
				seen[s.Key] = true
				e.Entries = append(e.Entries, s)
			}
			s = result.MEntry{Name: m[1]}
			continue
		}
		m = sRatingRe.FindStringSubmatch(line)
		if len(m) > 0 {
			log.Printf("SRating: %s", m[1])
			s.SRating, _ = strconv.Atoi(m[1])
			continue
		}
		m = LocaleRe.FindStringSubmatch(line)
		if s.SRating > 0 && len(m) > 0 {
			log.Printf("Locale: %s", m[1])
			s.Locale = m[1]
		}

		m = DescRe.FindStringSubmatch(line)
		if s.SRating > 0 && len(m) > 0 {
			log.Printf("Desc: %s", m[1])
			if s.Desc == "" {
				s.Desc = m[1]
			} else if len(strings.Split(s.Desc, " ")) < MaxDescWords {
				s.Desc = s.Desc + " " + m[1]
			}
			words := strings.Split(s.Desc, " ")
			if len(words) > MaxDescWords {
				s.Desc = strings.Join(words[0:MaxDescWords], " ") + " ..."
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
