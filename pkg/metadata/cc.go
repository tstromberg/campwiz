package metadata

import (
	"io/ioutil"

	"github.com/tstromberg/campwiz/pkg/relpath"
	"k8s.io/klog/v2"

	"gopkg.in/yaml.v2"
)

// LoadCC returns CC cross-reference data
func LoadCC() (map[string]XRef, error) {
	p := relpath.Find("metadata/cc.yaml")
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
		e.Source = ccd.Source
		xs[e.ID] = e
	}
	return xs, nil
}
