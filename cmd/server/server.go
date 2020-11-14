// The "main" package contains the command-line utility functions.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"text/template"

	"github.com/tstromberg/campwiz/pkg/provider"
	"github.com/tstromberg/campwiz/pkg/relpath"
	"k8s.io/klog/v2"
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
	Query   provider.Query
	Results []provider.Result
	Form    formValues
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Incoming request: %+v", r)
	klog.Infof("Incoming request: %+v", r)
	crit := provider.Query{}
	results, err := provider.Search(crit)
	klog.V(1).Infof("RESULTS: %+v", results)
	if err != nil {
		klog.Errorf("Search error: %s", err)
	}

	p := relpath.Find("templates/http.tmpl")
	outTmpl, err := ioutil.ReadFile(p)
	if err != nil {
		klog.Errorf("Failed to read template: %v", err)
	}
	tmpl := template.Must(template.New("http").Parse(string(outTmpl)))
	ctx := templateContext{
		Query:   crit,
		Results: results,
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
	http.HandleFunc("/", handler)
	klog.Fatal(http.ListenAndServe(":8080", nil))
}
