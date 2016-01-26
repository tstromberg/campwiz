// The "main" package contains the command-line utility functions.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"text/template"
	"time"

	"github.com/tstromberg/autocamper/data"
	"github.com/tstromberg/autocamper/query"
	"github.com/tstromberg/autocamper/result"
)

var (
	lat         = flag.Float64("lat", -122.4194155, "Base latitude")
	lon         = flag.Float64("lon", 37.7749295, "Base longitude")
	date        = flag.String("date", "", "Date to search for (YYYY-MM-DD). Defaults to 60 days from now.")
	nights      = flag.Int("nights", 2, "Number of nights to search for")
	maxPages    = flag.Int("max_pages", 10, "Number of pages to request")
	maxDistance = flag.Int("max_distance", 200, "Furthest distance in miles to query for")
	group       = flag.Bool("group", false, "Search for group sites")
	boat        = flag.Bool("boat", false, "Search for boat sites")
	walkin      = flag.Bool("walkin", true, "Search for walk-in sites")
	standard    = flag.Bool("standard", true, "Search for standard camp sites")
)

func main() {
	var t time.Time
	var err error
	flag.Parse()

	tmpl, err := template.ParseFiles("ascii.tmpl")
	if err != nil {
		panic(err)
	}

	if *date != "" {
		t, err = time.Parse("2006-01-02", *date)
		if err != nil {
			panic(fmt.Sprintf("Bad date: %s", *date))
		}
	} else {
		// Friday, two months from now.
		t = time.Now().AddDate(0, 2, 6-int(time.Now().Weekday())+2)
	}

	crit := query.Criteria{
		Lat:         *lat,
		Lon:         *lon,
		Date:        t,
		Nights:      *nights,
		MaxPages:    *maxPages,
		MaxDistance: *maxDistance,
	}
	err = data.LoadM("../../data/m.yaml")
	if err != nil {
		log.Fatalf("Could not load m.yaml: %v", err)
	}

	results, err := query.Search(crit)
	if err != nil {
		log.Fatalf("Search error: %s", err)
	}

	var filtered result.Results

	for _, r := range results {
		data.Merge(&r)

		if *group && r.Availability.Group > 0 {
			filtered = append(filtered, r)
			log.Printf("* (Group) %+v\n", r)
			continue
		}
		if *boat && r.Availability.Boat > 0 {
			filtered = append(filtered, r)
			log.Printf("* (Boat) %+v\n", r)
			continue
		}
		if *walkin && r.Availability.WalkIn > 0 {
			filtered = append(filtered, r)
			log.Printf("* (Walk-In) %+v\n", r)
			continue
		}
		if *standard && r.Availability.Standard > 0 {
			filtered = append(filtered, r)
			log.Printf("* (Standard) %+v\n", r)
			continue
		}

	}
	sort.Sort(filtered)
	err = tmpl.ExecuteTemplate(os.Stdout, "ascii.tmpl", filtered)
	if err != nil {
		panic(err)
	}
}
