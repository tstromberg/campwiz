// The "main" package contains the command-line utility functions.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"
	"time"

	"github.com/tstromberg/campwiz/pkg/cache"
	"github.com/tstromberg/campwiz/pkg/metadata"
	"github.com/tstromberg/campwiz/pkg/mix"
	"github.com/tstromberg/campwiz/pkg/search"
	"k8s.io/klog/v2"
)

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
		Dates:      []time.Time{time.Now().Add(24 * 120 * time.Hour)},
		StayLength: 4,
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
	flag.Parse()
	if err := processFlags(); err != nil {
		klog.Exitf("processing error: %v", err)
	}
}
