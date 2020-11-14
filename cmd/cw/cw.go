// The "main" package contains the command-line utility functions.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"
	"time"

	"github.com/tstromberg/campwiz/pkg/engine"
	"github.com/tstromberg/campwiz/pkg/mixer"
	"k8s.io/klog/v2"
)

const dateFormat = "2006-01-02"

type templateContext struct {
	Query engine.Query
	Results  []mixer.Result
	Form     formValues
}

func processFlags() error {
	q := engine.Query{
		Lon:    -122.07237049999999,
		Lat:    37.4092297,
		Dates:  []time.Time{time.Now().Add(24 * 90 * time.Hour)},
		Nights: 4,
	}

	rs, err := engine.Search(q)
	if err != nil {
		return fmt.Errorf("search: %w", err)
	}

	ms, err := mixer.Mix(rs, meta)
	klog.V(1).Infof("RESULTS: %+v", results)



	
	bs, err := ioutil.ReadFile("templates/ascii.tmpl")
	if err != nil {
		return fmt.Errorf("readfile: %w", err)
	}

	tmpl := template.Must(template.New("ascii").Parse(string(bs)))
	c := templateContext{
		Query: q,
		Results:  ms,
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
