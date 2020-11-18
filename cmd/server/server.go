// The "main" package contains the command-line utility functions.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"text/template"

	"github.com/tstromberg/campwiz/pkg/cache"
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
	Query   search.Query
	Results []search.Result
	Form    formValues
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Incoming request: %+v", r)
	klog.Infof("Incoming request: %+v", r)
	q := search.Query{}

	xrefs, err := metadata.Load()
	if err != nil {
		klog.Errorf("loadcc failed: %w", err)
	}

	rs, err := search.All(q, cs, xrefs)
	if err != nil {
		klog.Errorf("search: %w", err)
	}

	p := relpath.Find("templates/http.tmpl")
	outTmpl, err := ioutil.ReadFile(p)
	if err != nil {
		klog.Errorf("Failed to read template: %v", err)
	}
	tmpl := template.Must(template.New("http").Parse(string(outTmpl)))
	ctx := templateContext{
		Query:   q,
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
