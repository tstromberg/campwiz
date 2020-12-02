// import_cc imports data from exported CC HTML files
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/tstromberg/campwiz/pkg/campwiz"
	"github.com/tstromberg/campwiz/pkg/metadata"
	"github.com/tstromberg/campwiz/pkg/metasrc"
	"gopkg.in/yaml.v3"
	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	rf := campwiz.RefFile{}
	props := map[string]*campwiz.Property{}

	for _, path := range flag.Args() {
		f, err := os.Open(path)
		if err != nil {
			log.Fatal(err)
		}

		err = metasrc.CC(f, props)
		if err != nil {
			log.Fatalf("parse error: %v", err)
		}
	}

	for _, p := range props {
		for x, cg := range p.Campgrounds {
			for y, ref := range cg.Refs {
				p.Campgrounds[x].Refs[y].Desc = metadata.Compress(ref.Desc)
			}
		}
		rf.Properties = append(rf.Properties, p)
	}

	sort.Slice(rf.Properties, func(i, j int) bool { return rf.Properties[i].ID < rf.Properties[j].ID })

	d, err := yaml.Marshal(&rf)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	fmt.Printf("%s", d)
}
