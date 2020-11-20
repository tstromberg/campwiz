// The "main" package contains the command-line utility functions.
package main

import (
	goflag "flag"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"
	"time"

	pflag "github.com/spf13/pflag"
	"github.com/tstromberg/campwiz/pkg/cache"
	"github.com/tstromberg/campwiz/pkg/metadata"
	"github.com/tstromberg/campwiz/pkg/mix"
	"github.com/tstromberg/campwiz/pkg/search"
	"k8s.io/klog/v2"
)

var datesFlag *[]string = pflag.StringSlice("dates", []string{}, "dates to search for")

const dateFormat = "2006-01-02"

type templateContext struct {
	Query        search.Query
	MixedResults []mix.MixedResult
}

func processFlags() error {
	cs, err := cache.Initialize()
	if err != nil {
		return err
	}

	q := search.Query{
		Lon:        -122.07237049999999,
		Lat:        37.4092297,
		StayLength: 4,
	}

	for _, ds := range *datesFlag {
		t, err := time.Parse(dateFormat, ds)
		if err != nil {
			klog.Fatalf("unable to parse date %q: %v", ds, err)
		}
		q.Dates = append(q.Dates, t)
	}

	xrefs, err := metadata.Load()
	if err != nil {
		return fmt.Errorf("loadcc failed: %w", err)
	}

	rs, err := search.All(q, cs, xrefs)
	if err != nil {
		return fmt.Errorf("search: %w", err)
	}

	ms := mix.Combine(rs, xrefs)
	klog.V(1).Infof("RESULTS: %+v", ms)

	bs, err := ioutil.ReadFile("templates/ascii.tmpl")
	if err != nil {
		return fmt.Errorf("readfile: %w", err)
	}

	tmpl := template.Must(template.New("ascii").Parse(string(bs)))
	c := templateContext{
		Query:        q,
		MixedResults: ms,
	}

	err = tmpl.ExecuteTemplate(os.Stdout, "ascii", c)
	return err
}

func main() {
	//	wordPtr := flag.String("word", "foo", "a string")
	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	pflag.Parse()

	if err := processFlags(); err != nil {
		klog.Exitf("processing error: %v", err)
	}
}
