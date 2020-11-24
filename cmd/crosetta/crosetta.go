// crosetta maintains the camping rosetta stone file
package main

import (
	"flag"
	"fmt"
	"html"
	"log"
	"sort"
	"strings"

	"github.com/tstromberg/campwiz/pkg/cache"
	"github.com/tstromberg/campwiz/pkg/mangle"
	"github.com/tstromberg/campwiz/pkg/metadata"
	"github.com/tstromberg/campwiz/pkg/metasrc"
	"gopkg.in/yaml.v3"
	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	xrefs, err := metadata.Load()
	if err != nil {
		klog.Fatalf("xrefs: %s", err)
	}

	records := []metasrc.RosettaEntry{}

	// sort ids for predictable runtime behavior
	ids := []string{}
	for id := range xrefs {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		xref := xrefs[id]
		if !strings.HasPrefix(id, "/cc") {
			continue
		}
		nid := strings.Replace(id, "/cc", "/ca", 1)
		name := html.UnescapeString(xref.Name)
		locale := mangle.ShortLocale(html.UnescapeString(xref.Locale))

		refs := map[string]string{"CC": fmt.Sprintf("%s (%s)", name, locale)}

		records = append(records, metasrc.RosettaEntry{
			ID:          nid,
			ReserveName: name,
			Locale:      mangle.Locale(locale),
			Refs:        refs,
		})
	}

	cs, err := cache.Initialize()
	if err != nil {
		klog.Fatalf("cache init: %v", err)
	}

	for i, r := range records {
		ur, err := metasrc.RosettaSearch(r, cs)
		if err != nil {
			klog.Errorf("search error: %v", err)
			continue
		}
		records[i] = ur

	}

	sort.Slice(records, func(i, j int) bool { return records[i].ID < records[j].ID })

	content, err := yaml.Marshal(&records)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	fmt.Printf("%s\n", content)
}
