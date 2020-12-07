package metadata

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/tstromberg/campwiz/pkg/campwiz"
	"github.com/tstromberg/campwiz/pkg/relpath"
	"k8s.io/klog/v2"

	"gopkg.in/yaml.v3"
)

var (
	CompressHeader = `H4sIAAAAAAAA/`
	CompressPrefix = `z`
)

// LoadAll returns all cross-reference data
func LoadAll() (map[string]campwiz.Source, map[string]*campwiz.Property, error) {
	csrcs := map[string]campwiz.Source{}
	cprops := map[string]*campwiz.Property{}

	for _, p := range []string{"metadata/srcs.yaml", "metadata/ca.yaml"} {
		path := relpath.Find(p)
		if path == "" {
			klog.Errorf("unable to find %s", p)
		}
		klog.Infof("%s is at %s", p, path)

		srcs, props, err := loadPath(path)
		if err != nil {
			return srcs, props, fmt.Errorf("loadpath %q: %v", path, err)
		}

		for k, v := range srcs {
			csrcs[k] = v
		}
		for k, v := range props {
			cprops[k] = v
		}
	}

	return csrcs, cprops, nil
}

// LoadPath loads YAML data from a Path!
// LoadCC returns CC cross-reference data
func loadPath(path string) (map[string]campwiz.Source, map[string]*campwiz.Property, error) {
	p := relpath.Find(path)
	f, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, nil, err
	}

	var ccd campwiz.RefFile
	err = yaml.Unmarshal(f, &ccd)
	if err != nil {
		return nil, nil, err
	}

	klog.V(1).Infof("Loaded %d entries from %s ...", len(ccd.Properties), p)

	props := map[string]*campwiz.Property{}
	for _, p := range ccd.Properties {
		props[p.ID] = p
	}
	return ccd.Sources, props, nil
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
