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

	"github.com/tstromberg/autocamper/data"
	"gopkg.in/yaml.v2"
)

var (
	titleRe      = regexp.MustCompile(`^\d+ ([A-Z].*)`)
	sRatingRe    = regexp.MustCompile(`^Scenic rating: (\d+)`)
	DescRe       = regexp.MustCompile(`^([A-Z].*[\.\!\)])$`)
	MaxDescWords = 50
)

func parse(scanner *bufio.Scanner) data.MEntries {
	var e data.MEntries
	s := data.MEntry{}
	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("Line: %s", line)
		m := titleRe.FindStringSubmatch(line)
		if len(m) > 0 && line == strings.ToUpper(line) {
			log.Printf("Title: %s", m[1])
			// Clear the previous entry.
			if s.Name != "" && s.SRating > 0 {
				e.Entries = append(e.Entries, s)
			}
			s = data.MEntry{Name: m[1]}
			continue
		}
		m = sRatingRe.FindStringSubmatch(line)
		if len(m) > 0 {
			log.Printf("SRating: %s", m[1])
			s.SRating, _ = strconv.Atoi(m[1])
			continue
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
	e.Entries = append(e.Entries, s)
	return e
}

func main() {
	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	es := parse(scanner)
	d, err := yaml.Marshal(&es)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("%s", d)

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
