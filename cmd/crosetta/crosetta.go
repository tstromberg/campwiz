// crosetta maintains the camping rosetta stone file
package main

import (
	"flag"
	"fmt"
	"html"
	"log"
	"sort"
	"strings"

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

	for id, xref := range xrefs {
		if !strings.HasPrefix(id, "/cc") {
			continue
		}
		nid := strings.Replace(id, "/cc", "/ca", 1)
		name := html.UnescapeString(xref.Name)
		locale := mangle.ShortLocale(html.UnescapeString(xref.Locale))

		records = append(records, metasrc.RosettaEntry{
			ID:              nid,
			ReservationName: name,
			Locale:          mangle.Locale(locale),
			Stienstra:       fmt.Sprintf("%s (%s)", name, locale),
		})
	}
	sort.Slice(records, func(i, j int) bool { return records[i].ID < records[j].ID })

	content, err := yaml.Marshal(&records)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	fmt.Printf("%s\n", content)
}
