package metasrc

import (
	"fmt"
	"net/url"

	"github.com/tstromberg/campwiz/pkg/cache"
)

var nonAuthHost = map[string]bool{
	"californiasbestcamping.com": true,
	"www.yelp.com":               true,
	"www.dyrt.com":               true,
	"www.tripadvisor.com":        true,
}

type RosettaEntry struct {
	ID          string `yaml:"id,omitempty"`
	URL         string `yaml:"url,omitempty"`
	ReserveName string `yaml:"reserve_name,omitempty"`
	Locale      string `yaml:"locale,omitempty"`

	Refs map[string]string `yaml:"refs"`
}

func RosettaSearch(e RosettaEntry, cs cache.Store) (RosettaEntry, error) {
	s := e.Refs["CC"]
	ba, err := BingSearch(cs, fmt.Sprintf("camping %s", s))
	if err != nil {
		return e, fmt.Errorf("bing search: %w", err)
	}

	for _, b := range ba.WebPages.Value {
		u, err := url.Parse(b.URL)
		if err != nil {
			return e, fmt.Errorf("url parse: %v", err)
		}

		host := u.Hostname()
		if e.Refs[host] == "" {
			e.Refs[host] = b.URL
		}

		if e.URL == "" && !nonAuthHost[host] {
			e.URL = b.URL
		}

	}
	return e, nil
}
