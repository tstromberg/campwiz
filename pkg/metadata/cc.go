package metadata

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"strings"

	"github.com/tstromberg/campwiz/pkg/relpath"
	"k8s.io/klog/v2"

	"gopkg.in/yaml.v2"
)

// LoadCC returns CC cross-reference data
func LoadCC() (map[string]XRef, error) {
	p := relpath.Find("metadata/imported/cc.yaml")
	f, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}

	var ccd XrefData
	err = yaml.Unmarshal(f, &ccd)
	if err != nil {
		return nil, err
	}

	klog.V(1).Infof("Loaded %d entries from %s ...", len(ccd.Entries), p)

	xs := map[string]XRef{}
	for _, e := range ccd.Entries {
		if strings.HasPrefix(e.Desc, CompressPrefix) {
			e.Desc = decompress(e.Desc)
		}
		e.Source = ccd.Source
		xs[e.ID] = e
	}
	return xs, nil
}

func decompress(s string) string {
	bs, err := base64.RawStdEncoding.DecodeString(CompressHeader + s[1:])
	if err != nil {
		klog.Fatalf("decode fail: %v", err)
	}
	klog.Infof("bs: %s", strings.Replace(s, CompressPrefix, CompressHeader, 1))

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
