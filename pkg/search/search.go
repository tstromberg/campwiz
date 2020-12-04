package search

import (
	"fmt"
	"sort"

	"github.com/tstromberg/campwiz/pkg/backend"
	"github.com/tstromberg/campwiz/pkg/cache"
	"github.com/tstromberg/campwiz/pkg/campwiz"
	"k8s.io/klog"
)

var DefaultProviders = []string{"ramerica", "rcalifornia", "scc", "smc"}

// Run is a one-stop query shop: talks to backends, annotates, provides filtering
func Run(providers []string, q campwiz.Query, cs cache.Store, props map[string]*campwiz.Property) ([]campwiz.Result, []error) {
	rs, errs := unfiltered(providers, q, cs)

	as := []campwiz.Result{}
	for _, r := range rs {
		as = append(as, annotate(r, props))
	}

	fs := filter(q, as)

	sort.Slice(fs, func(i, j int) bool { return fs[i].Rating > fs[j].Rating })
	return fs, errs
}

// unfiltered searches for results across providers, without filters
func unfiltered(providers []string, q campwiz.Query, cs cache.Store) ([]campwiz.Result, []error) {
	klog.V(1).Infof("search campwiz.Query: %+v", q)

	results := []campwiz.Result{}
	errs := []error{}

	// There is an opportunity to parallelize this with channels if anyone is keen to do so
	for _, pname := range providers {
		p, err := backend.New(backend.Config{Type: pname, Store: cs})
		if err != nil {
			errs = append(errs, fmt.Errorf("%s init: %v", pname, err))
			continue
		}

		prs, err := p.List(q)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s list: %v", pname, err))
			continue
		}

		results = append(results, prs...)
	}

	return results, errs
}
