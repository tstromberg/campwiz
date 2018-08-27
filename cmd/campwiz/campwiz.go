// The "main" package contains the command-line utility functions.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/golang/glog"
	"github.com/tstromberg/campwiz/data"
	"github.com/tstromberg/campwiz/query"
	"github.com/tstromberg/campwiz/result"
)

var (
	lat         = flag.Float64("lat", -122.4194155, "Base latitude")
	lon         = flag.Float64("lon", 37.7749295, "Base longitude")
	dates       = flag.String("dates", "", "Dates to search for (YYYY-MM-DD). Defaults to 60 days from now.")
	nights      = flag.Int("nights", 2, "Number of nights to search for")
	maxPages    = flag.Int("max_pages", 10, "Number of pages to request")
	maxDistance = flag.Int("max_distance", 150, "Furthest distance in miles to query for")
	group       = flag.Bool("group", false, "Search for group sites")
	boat        = flag.Bool("boat", false, "Search for boat sites")
	walkin      = flag.Bool("walkin", true, "Search for walk-in sites")
	standard    = flag.Bool("standard", true, "Search for standard camp sites")
	verbose     = flag.Bool("v", false, "Enable verbose mode")
)

type TemplateContext struct {
	Criteria query.Criteria
	Results  result.Results
}

func main() {
	var t time.Time
	var err error
	flag.Parse()
	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}
	tmpl, err := template.ParseFiles("ascii.tmpl")
	if err != nil {
		panic(err)
	}
	err = data.LoadM("../../data/m.yaml")
	if err != nil {
		panic(fmt.Sprintf("Could not load m.yaml: %v", err))
	}

	var parsedDates []time.Time
	if *dates == "" {
		// Default to a Friday two months from now.
		parsedDates = append(parsedDates, time.Now().AddDate(0, 2, 6-int(time.Now().Weekday())+2))
	} else {
		for _, d := range strings.Split(*dates, ",") {
			if d != "" {
				t, err = time.Parse("2006-01-02", d)
				if err != nil {
					panic(fmt.Sprintf("Bad date: %s", d))
				}
				parsedDates = append(parsedDates, t)
			}
		}
	}

	// For each date, output the template.
	crit := query.Criteria{
		Lat:             *lat,
		Lon:             *lon,
		Dates:           parsedDates,
		Nights:          *nights,
		MaxPages:        *maxPages,
		MaxDistance:     *maxDistance,
		IncludeStandard: *standard,
		IncludeGroup:    *group,
		IncludeBoatIn:   *boat,
		IncludeWalkIn:   *walkin,
	}
	fmt.Printf("Criteria: %+v\n", crit)
	fmt.Println("Searching (this may take a minute) ...")
	results, err := query.Search(crit)
	glog.V(1).Infof("RESULTS: %+v", results)
	if err != nil {
		panic(fmt.Sprintf("Search error: %s", err))
	}
	ctx := TemplateContext{Criteria: crit, Results: results}
	err = tmpl.ExecuteTemplate(os.Stdout, "ascii.tmpl", ctx)
	if err != nil {
		panic(err)
	}
}
