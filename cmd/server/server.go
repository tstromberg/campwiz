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

type TemplateContext struct {
	Criteria query.Criteria
	Results  result.Results
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
	ctx := TemplateContext{Criteria: crit, Results: results}
	tmpl.ExecuteTemplate(w, "http", ctx)
}

func init() {
	flag.Parse()
}
func main() {
	http.HandleFunc("/", handler)
	glog.Fatal(http.ListenAndServe(":8080", nil))
}
