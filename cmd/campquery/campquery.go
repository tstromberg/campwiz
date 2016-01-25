// The "main" package contains the command-line utility functions.
package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/tstromberg/autocamper/data"
	"github.com/tstromberg/autocamper/query"
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

	for _, r := range results {
		data.Merge(&r)

		if *group && r.Availability.Group > 0 {
			fmt.Printf("* (Group) %+v\n", r)
			continue
		}
		if *boat && r.Availability.Boat > 0 {
			fmt.Printf("* (Boat) %+v\n", r)
			continue
		}
		if *walkin && r.Availability.WalkIn > 0 {
			fmt.Printf("* (Walk-In) %+v\n", r)
			continue
		}
		if *standard && r.Availability.Standard > 0 {
			fmt.Printf("* (Standard) %+v\n", r)
			continue
		}
	}
}
