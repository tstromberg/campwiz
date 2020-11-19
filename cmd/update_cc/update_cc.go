// update_cc updates CC metadata (cc.yaml) from exported HTML files
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tstromberg/campwiz/pkg/metadata"
	"github.com/tstromberg/campwiz/pkg/parse"
	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	xd := metadata.XrefData{
		Source: metadata.XrefSource{
			Name:       "CC",
			RatingDesc: "Scenery",
			RatingMax:  10,
		},
	}

	for _, path := range flag.Args() {
		f, err := os.Open(path)
		if err != nil {
			log.Fatal(err)
		}

		xrefs, err := parse.CC(f)
		if err != nil {
			log.Fatalf("parse error: %v", err)
		}

		xd.Entries = append(xd.Entries, xrefs...)
	}

	d, err := yaml.Marshal(&xd)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("%s", d)
}
