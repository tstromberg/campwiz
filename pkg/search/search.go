package search

import (
	"fmt"

	"github.com/tstromberg/campwiz/pkg/backend"
	"github.com/tstromberg/campwiz/pkg/cache"
	"github.com/tstromberg/campwiz/pkg/campwiz"
	"k8s.io/klog"
)

var (
	DefaultProviders = []string{"ramerica", "rcalifornia", "scc", "smc"}
)

// Run is a one-stop query shop: talks to backends, annotates, provides filtering
func Run(providers []string, q campwiz.Query, cs cache.Store, refs map[string]campwiz.Ref) ([]campwiz.AnnotatedResult, []error) {
	rs, errs := unfiltered(providers, q, cs)
	return annotate(rs, refs), errs
}

// unfiltered searches for results across providers, without filters
func unfiltered(providers []string, q campwiz.Query, cs cache.Store) ([]campwiz.Result, []error) {
	klog.Infof("search campwiz.Query: %+v", q)

	// TODO: Paralellize
	results := []campwiz.Result{}
	errs := []error{}

	for _, pname := range providers {
		p, err := backend.New(backend.Config{Type: pname, Store: cs})
		if err != nil {
			errs = append(errs, fmt.Errorf("%s init: %v", pname, err))
			continue
		}

		prs, err := p.List(q)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s list: %v", pname, err))
			klog.Errorf("%s list failed: %v", err)
			continue
		}

		results = append(results, prs...)
	}

	return results, errs
}
