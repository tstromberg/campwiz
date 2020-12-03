// The "main" package contains the command-line utility functions.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"text/template"

	"github.com/tstromberg/campwiz/pkg/cache"
	"github.com/tstromberg/campwiz/pkg/campwiz"
	"github.com/tstromberg/campwiz/pkg/metadata"
	"github.com/tstromberg/campwiz/pkg/relpath"
	"github.com/tstromberg/campwiz/pkg/search"
	"k8s.io/klog/v2"
)

var (
	cs cache.Store
)

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
	Query   campwiz.Query
	Results []campwiz.Result
	Sources map[string]campwiz.Source
	Form    formValues
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Incoming request: %+v", r)
	klog.Infof("Incoming request: %+v", r)
	q := campwiz.Query{}

	srcs, props, err := metadata.LoadAll()
	if err != nil {
		klog.Errorf("loadall failed: %w", err)
	}

	rs, errs := search.Run(search.DefaultProviders, q, cs, props)
	if errs != nil {
		klog.Errorf("search: %v", errs)
	}

	p := relpath.Find("templates/http.tmpl")
	outTmpl, err := ioutil.ReadFile(p)
	if err != nil {
		klog.Errorf("Failed to read template: %v", err)
	}
	tmpl := template.Must(template.New("http").Parse(string(outTmpl)))
	ctx := templateContext{
		Query:   q,
		Sources: srcs,
		Results: rs,
		Form: formValues{
			Dates: "2018-09-20",
		},
	}
	err = tmpl.ExecuteTemplate(w, "http", ctx)
	if err != nil {
		klog.Errorf("template: %v", err)
	}
}

func init() {
	flag.Parse()
}

func main() {
	var err error
	cs, err = cache.Initialize()
	if err != nil {
		klog.Exitf("error: %w", err)
	}

	http.HandleFunc("/", handler)
	klog.Fatal(http.ListenAndServe(":8080", nil))
}
