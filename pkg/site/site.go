// Package site define HTTP handlers.
package site

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/tstromberg/campwiz/pkg/cache"
	"github.com/tstromberg/campwiz/pkg/campwiz"
	"k8s.io/klog/v2"
)

// VERSION is what version of campwiz we advertise as.
const VERSION = "v1.0.0-DEV (master)"

// Config is how external users interact with this package.
type Config struct {
	BaseDirectory string
	Cache         cache.Store
	Sources       map[string]campwiz.Source
	Properties    map[string]*campwiz.Property
	Providers     []string

	// For hardcoding a site to a particular address
	Latitude  float64
	Longitude float64
}

func New(c *Config) *Handlers {
	return &Handlers{
		c:         *c,
		startTime: time.Now(),
	}
}

// Handlers manages local state
type Handlers struct {
	c Config

	startTime time.Time
}

// Root redirects to leaderboard.
func (h *Handlers) Root() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/search", http.StatusSeeOther)
	}
}

// Healthz returns a dummy healthz page - it's always happy here!
func (h *Handlers) Healthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("ok: started at %s", h.startTime)))
	}
}

// Threadz returns a threadz page
func (h *Handlers) Threadz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		klog.Infof("GET %s: %v", r.URL.Path, r.Header)
		w.WriteHeader(http.StatusOK)
		w.Write(stack())
	}
}

// stack returns a formatted stack trace of all goroutines
func stack() []byte {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			return buf[:n]
		}
		buf = make([]byte, 2*len(buf))
	}
}

func (h *Handlers) error(w http.ResponseWriter, err error) {
	msg := err.Error()
	_, file, line, ok := runtime.Caller(0)
	if ok {
		msg = fmt.Sprintf("%s:%d: %v", file, line, err)
	}
	klog.Errorf(msg)
	http.Error(w, msg, 500)
}
