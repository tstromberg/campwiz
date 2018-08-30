// The "main" package contains the command-line utility functions.
package main

import (
	"flag"
	"fmt"
	"net/http"
	"text/template"

	"github.com/golang/glog"
	"github.com/tstromberg/campwiz/data"
	"github.com/tstromberg/campwiz/query"
	"github.com/tstromberg/campwiz/result"
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

type TemplateContext struct {
	Criteria query.Criteria
	Results  result.Results
	Form     formValues
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Incoming request: %+v", r)
	glog.Infof("Incoming request: %+v", r)
	crit := query.Criteria{}
	results, err := query.Search(crit)
	glog.V(1).Infof("RESULTS: %+v", results)
	if err != nil {
		glog.Errorf("Search error: %s", err)
	}

	outTmpl, err := data.Read("http.tmpl")
	if err != nil {
		glog.Errorf("Failed to read template: %v", err)
	}
	tmpl := template.Must(template.New("http").Parse(string(outTmpl)))
	ctx := TemplateContext{
		Criteria: crit,
		Results:  results,
		Form: formValues{
			Dates: "2018-09-20",
		},
	}
	err = tmpl.ExecuteTemplate(w, "http", ctx)
	if err != nil {
		glog.Errorf("template: %v", err)
	}
}

func init() {
	flag.Parse()
}
func main() {
	http.HandleFunc("/", handler)
	glog.Fatal(http.ListenAndServe(":8080", nil))
}
