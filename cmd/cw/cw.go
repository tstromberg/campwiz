// The "main" package contains the command-line utility functions.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"
	"time"

	"github.com/tstromberg/campwiz/pkg/query"
	"github.com/tstromberg/campwiz/pkg/result"
	"k8s.io/klog/v2"
)

const dateFormat = "2006-01-02"

type formValues struct {
	Dates    string
	Nights   int
	Distance int
	Standard bool
	Group    bool
	WalkIn   bool
	BoatIn   bool
}

type templateContext struct {
	Criteria query.Criteria
	Results  result.Results
	Form     formValues
}

func processFlags() error {
	crit := query.Criteria{
		Lon:    -122.07237049999999,
		Lat:    37.4092297,
		Dates:  []time.Time{time.Now().Add(24 * 90 * time.Hour)},
		Nights: 4,
	}

	results, err := query.Search(crit)
	if err != nil {
		return fmt.Errorf("search: %w", err)
	}

	klog.V(1).Infof("RESULTS: %+v", results)

	bs, err := ioutil.ReadFile("../../templates/ascii.tmpl")
	if err != nil {
		return fmt.Errorf("readfile: %w", err)
	}

	tmpl := template.Must(template.New("ascii").Parse(string(bs)))
	c := templateContext{
		Criteria: crit,
		Results:  results,
		Form: formValues{
			Dates: "2018-09-20",
		},
	}

	err = tmpl.ExecuteTemplate(os.Stdout, "ascii", c)
	return err
}

func main() {
	//	wordPtr := flag.String("word", "foo", "a string")

	flag.Parse()
	klog.InitFlags(nil)
	if err := processFlags(); err != nil {
		klog.Exitf("processing error: %v", err)
	}
}
