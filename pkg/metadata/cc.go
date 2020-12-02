package metadata

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"strings"

	"github.com/tstromberg/campwiz/pkg/campwiz"
	"github.com/tstromberg/campwiz/pkg/relpath"
	"k8s.io/klog/v2"

	"gopkg.in/yaml.v3"
)

// LoadCC returns CC cross-reference data
func LoadCC() (map[string]*campwiz.Property, error) {
	p := relpath.Find("metadata/ca.yaml")
	f, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}

	var ccd campwiz.RefFile
	err = yaml.Unmarshal(f, &ccd)
	if err != nil {
		return nil, err
	}

	klog.V(1).Infof("Loaded %d entries from %s ...", len(ccd.Properties), p)

	props := map[string]*campwiz.Property{}
	for _, p := range ccd.Properties {
		props[p.ID] = p
	}
	return props, nil

	/*
		xs := map[string]campwiz.Ref{}
		for _, e := range ccd.Properties {
			if strings.HasPrefix(e.Desc, CompressPrefix) {
				e.Desc = Decompress(e.Desc)
			}
			e.Source = ccd.Source
			xs[e.ID] = e
		}
	*/
}

func Decompress(s string) string {
	bs, err := base64.RawStdEncoding.DecodeString(CompressHeader + s[1:])
	if err != nil {
		klog.Fatalf("decode fail: %v", err)
	}

	buf := bytes.NewReader(bs)
	zr, err := gzip.NewReader(buf)
	if err != nil {
		klog.Fatalf("reader fail: %v", err)
	}

	d, err := ioutil.ReadAll(zr)
	if err != nil {
		klog.Fatalf("read fail: %v", err)
	}

	return string(d)
}

func Compress(s string) string {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err := zw.Write([]byte(s))
	if err != nil {
		klog.Fatalf("write error: %v", err)
	}

	if err := zw.Close(); err != nil {
		klog.Fatalf("close error: %v", err)
	}

	return strings.Replace(base64.RawStdEncoding.EncodeToString(buf.Bytes()), CompressHeader, CompressPrefix, -1)
}
