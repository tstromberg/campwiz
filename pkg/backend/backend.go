package backend

import (
	"fmt"
	"net/http/cookiejar"
	"time"

	"github.com/tstromberg/campwiz/pkg/cache"
	"github.com/tstromberg/campwiz/pkg/campwiz"
	"k8s.io/klog/v2"
)

var (
	// searchPageExpiry is how long search pages can be cached for.
	searchPageExpiry = time.Duration(6*3600) * time.Second

	// amount of time to sleep between uncached fetches
	uncachedDelay = time.Millisecond * 600

	// maximum number of pages to fetch
	maxPages = 15
)

// Provider is a common interface for backend providers
type Provider interface {
	// Name is a human readable name for a runtime
	Name() string

	// List lists open campsites
	List(q campwiz.Query) ([]campwiz.Result, error)
}

// Config is runtime configuration
type Config struct {
	// Type of backend to use
	Type string
	// Store is the cache implementation to use
	Store cache.Store
}

// New returns an appropriately configured backend
func New(c Config) (Provider, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("cookiejar: %w", err)
	}

	switch c.Type {
	case "ramerica":
		return &RAmerica{store: c.Store, jar: jar}, nil
	case "rcalifornia":
		return &RCalifornia{store: c.Store, jar: jar}, nil
	case "scc":
		return &SantaClaraCounty{store: c.Store, jar: jar}, nil
	case "smc":
		return &SanMateoCounty{store: c.Store, jar: jar}, nil
	default:
		return nil, fmt.Errorf("unknown backend type: %q", c.Type)
	}
}

// mergeDates merges multiple dates together
func mergeDates(res []campwiz.Result) []campwiz.Result {
	klog.V(1).Infof("Merging %d results ...", len(res))
	m := make(map[string]campwiz.Result)
	for _, r := range res {
		key := r.ResURL + r.ResID // Merge campwiz.Availability metadata.
		if val, exists := m[key]; exists {
			klog.V(1).Infof("%s: Appending Availability: %+v (previous: %+v)", key, r.Availability, val.Availability)
			val.Availability = append(val.Availability, r.Availability...)
			// map items are immutable.
			m[key] = val
			klog.V(1).Infof("%s campwiz.Availability now: %+v", key, m[key].Availability)
		} else {
			klog.V(1).Infof("%s: Not yet seen: %+v", key, r)
			m[key] = r
		}
	}

	var merged []campwiz.Result
	for k, v := range m {
		klog.V(1).Infof("%s: %+v", k, v)
		merged = append(merged, v)
	}
	return merged
}
