package site

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"text/template"
	"time"

	"github.com/tstromberg/campwiz/pkg/campwiz"
	"github.com/tstromberg/campwiz/pkg/mangle"
	"github.com/tstromberg/campwiz/pkg/search"
	"k8s.io/klog/v2"
)

type templateContext struct {
	Query   campwiz.Query
	Results []campwiz.Result
	Sources map[string]campwiz.Source
	Errors  []error

	Today      time.Time
	SelectDate time.Time
	Version    string
}

func futureFriday() time.Time {
	try := time.Now().Add(time.Duration(7 * 6 * 24 * time.Hour))
	offset := 5 - int(try.Weekday())
	return try.Add(time.Duration(offset) * 24 * time.Hour)
}

// Search returns search results
func (h *Handlers) Search() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		klog.Infof("Incoming request: %+v", r)
		q := campwiz.Query{
			Lon:         h.c.Longitude,
			Lat:         h.c.Latitude,
			StayLength:  getInt(r.URL, "nights", 2),
			MaxDistance: getInt(r.URL, "distance", 100),
			MinRating:   getFloat(r.URL, "min_rating", 0.0),
			Keywords:    []string{getStr(r.URL, "keywords", "")},
		}

		selectDate := futureFriday()

		for _, ds := range r.URL.Query()["dates"] {
			t, err := time.Parse("2006-01-02", ds)
			if err != nil {
				h.error(w, err)
				return
			}
			q.Dates = append(q.Dates, t)
			selectDate = t
		}

		var rs []campwiz.Result
		var errs []error

		if len(q.Dates) > 0 {
			rs, errs = search.Run(h.c.Providers, q, h.c.Cache, h.c.Properties)
			if len(errs) > 0 {
				klog.Errorf("search errors: %v", errs)
			}
		}

		p := filepath.Join(h.c.BaseDirectory, "search.tmpl")
		outTmpl, err := ioutil.ReadFile(p)
		if err != nil {
			h.error(w, err)
			return
		}

		fmap := template.FuncMap{
			"Ellipsis": ellipse,
			"toDate":   toDate,
		}

		tmpl := template.Must(template.New("http").Funcs(fmap).Parse(string(outTmpl)))
		ctx := templateContext{
			Query:      q,
			Sources:    h.c.Sources,
			Results:    rs,
			Errors:     errs,
			SelectDate: selectDate,
			Today:      time.Now(),
			Version:    VERSION,
		}
		err = tmpl.ExecuteTemplate(w, "http", ctx)
		if err != nil {
			h.error(w, err)
			return
		}
	}
}

func ellipse(s string) string {
	return mangle.Ellipsis(s, 100)
}

func toDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// helper to get integers from a URL
func getInt(url *url.URL, key string, fallback int) int {
	vals := url.Query()[key]
	if len(vals) == 1 {
		i, err := strconv.ParseInt(vals[0], 10, 32)
		if err != nil {
			klog.Warningf("bad %s int value: %v", key, vals)
			return fallback
		}
		return int(i)
	}
	return fallback
}

// helper to get floats from a URL
func getFloat(url *url.URL, key string, fallback float64) float64 {
	vals := url.Query()[key]
	if len(vals) == 1 {
		i, err := strconv.ParseFloat(vals[0], 64)
		if err != nil {
			klog.Warningf("bad %s int value: %v", key, vals)
			return fallback
		}
		return i
	}
	return fallback
}

// helper to get string from a URL
func getStr(url *url.URL, key string, fallback string) string {
	vals := url.Query()[key]
	if len(vals) > 0 {
		return vals[0]
	}
	return fallback
}
