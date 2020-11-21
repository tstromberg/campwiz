// update_cc updates CC metadata (cc.yaml) from exported HTML files
package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/tstromberg/campwiz/pkg/campwiz"
	"github.com/tstromberg/campwiz/pkg/metadata"
	"github.com/tstromberg/campwiz/pkg/metasrc"
	"gopkg.in/yaml.v3"
	"k8s.io/klog/v2"
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	xd := campwiz.RefData{
		Source: campwiz.RefSource{
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

		xrefs, err := metasrc.CC(f)
		if err != nil {
			log.Fatalf("parse error: %v", err)
		}

		for _, x := range xrefs {
			x.Desc = compress(x.Desc)
			xd.Entries = append(xd.Entries, x)
		}
	}

	d, err := yaml.Marshal(&xd)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("%s", d)
}

func compress(s string) string {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err := zw.Write([]byte(s))
	if err != nil {
		klog.Fatalf("write error: %v", err)
	}

	if err := zw.Close(); err != nil {
		klog.Fatalf("close error: %v", err)
	}

	return strings.Replace(base64.RawStdEncoding.EncodeToString(buf.Bytes()), metadata.CompressHeader, metadata.CompressPrefix, -1)
}
