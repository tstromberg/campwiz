// The "main" package contains the command-line utility functions.
package main

import (
	"flag"
	goflag "flag"
	"fmt"
	"os"
	"text/template"
	"time"

	"github.com/mgutz/ansi"
	pflag "github.com/spf13/pflag"
	"github.com/tstromberg/campwiz/pkg/cache"
	"github.com/tstromberg/campwiz/pkg/campwiz"
	"github.com/tstromberg/campwiz/pkg/mangle"
	"github.com/tstromberg/campwiz/pkg/metadata"
	"github.com/tstromberg/campwiz/pkg/search"
	"k8s.io/klog/v2"
)

var (
	datesFlag       *[]string      = pflag.StringSlice("dates", []string{"2021-03-05"}, "dates to search for")
	milesFlag       *int           = pflag.Int("max_distance", 200, "distance to search within")
	nightsFlag      *int           = pflag.Int("nights", 2, "number of nights to stay")
	minRatingFlag   *float64       = pflag.Float64("min_rating", 0, "minimum scenery rating for inclusion")
	keywordsFlag    *[]string      = pflag.StringSlice("keywords", nil, "keywords to search for")
	maxCacheAgeFlag *time.Duration = pflag.Duration("max_cache_age", cache.RecommendedMaxAge, "max age of cache")
	latFlag         *float64       = pflag.Float64("lat", 37.4092297, "latitude to search from")
	lonFlag         *float64       = pflag.Float64("lon", -122.07237049999999, "longitude to search from")
	providersFlag   *[]string      = pflag.StringSlice("providers", search.DefaultProviders, "site providers to include")

	outTmpl = `
{{ $srcs := .Sources }}
{{ range $i, $r := .Results}}
{{ Color "(" "yellow+d" }}{{ printf "#%d" $i | yellow }}{{ Color ")" "yellow+d" }} {{ Color $r.Name "green+h" }} {{ Color "(" "black+h" }}{{ printf "%.0fmi" $r.Distance | green }}{{ with $r.Locale }}{{ Color "," "black+h"}} {{ . | green }}{{ end }}{{ Color ")" "black+h" }}
{{- range $r.Availability}}
{{ Color "  >" "cyan" }} {{ printf "%s %d"  .Date.Month .Date.Day | hwhite }}{{ Color ":" "cyan" }} {{.SpotCount}}x{{.Kind}} - {{.URL | cyan }}
{{- end }}
{{ with $r.KnownCampground }}
{{- range $k, $v := .Refs -}}
 {{- $src := index $srcs $k -}}
 {{ Color "  *" "magenta" }} {{ $src.Name | hmagenta }}: {{ printf "%.0f" $v.Rating | hwhite }}{{ Color "/" "black+h" }}{{ printf "%0.0f" $src.RatingMax | hwhite }}{{ with $src.RatingDesc }} {{$src.RatingDesc }}{{ end }}{{ with $v.Lists }}{{ Color ", " "black+h" }}{{ range . }} {{ printf "#%d" .Place | hmagenta }} {{ .Title | hwhite }}{{ end }}{{ end }}
{{- end }}
{{ end }}
  {{ with $r.Desc | Ellipsis }}{{ . }}{{ end }}
{{ end }}

{{- range .Errors}}{{ Color "ERROR: " "red" }}{{ printf "%s" . | yellow }}{{ end -}}
`
)

const dateFormat = "2006-01-02"

type templateContext struct {
	Query   campwiz.Query
	Sources map[string]campwiz.Source
	Results []campwiz.Result
	Errors  []error
}

func processFlags() error {
	cs, err := cache.New(cache.Config{MaxAge: *maxCacheAgeFlag})
	if err != nil {
		return err
	}

	q := campwiz.Query{
		Lon:         *lonFlag,
		Lat:         *latFlag,
		StayLength:  *nightsFlag,
		MaxDistance: *milesFlag,
		MinRating:   *minRatingFlag,
		Keywords:    *keywordsFlag,
	}

	for _, ds := range *datesFlag {
		t, err := time.Parse(dateFormat, ds)
		if err != nil {
			klog.Fatalf("unable to parse date %q: %v", ds, err)
		}
		q.Dates = append(q.Dates, t)
	}

	srcs, props, err := metadata.LoadAll()
	if err != nil {
		return fmt.Errorf("loadall failed: %w", err)
	}

	ms, errs := search.Run(*providersFlag, q, cs, props)

	fmap := template.FuncMap{
		"Ellipsis": ellipse,
		"Color":    ansi.Color,
		"yellow":   func(s string) string { return ansi.Color(s, "yellow") },
		"green":    func(s string) string { return ansi.Color(s, "green") },
		"cyan":     func(s string) string { return ansi.Color(s, "cyan") },
		"blue":     func(s string) string { return ansi.Color(s, "blue") },
		"magenta":  func(s string) string { return ansi.Color(s, "magenta") },
		"hyellow":  func(s string) string { return ansi.Color(s, "yellow+h") },
		"hgreen":   func(s string) string { return ansi.Color(s, "green+h") },
		"hblue":    func(s string) string { return ansi.Color(s, "blue+h") },
		"hmagenta": func(s string) string { return ansi.Color(s, "magenta+h") },
		"hwhite":   func(s string) string { return ansi.Color(s, "white+h") },
		"grey":     func(s string) string { return ansi.Color(s, "black+h") },
	}

	t := template.Must(template.New("ascii").Funcs(fmap).Parse(outTmpl))

	c := templateContext{
		Query:   q,
		Results: ms,
		Sources: srcs,
		Errors:  errs,
	}

	err = t.ExecuteTemplate(os.Stdout, "ascii", c)
	return err
}

func ellipse(s string) string {
	return mangle.Ellipsis(s, 100)
}

func main() {
	//	wordPtr := flag.String("word", "foo", "a string")
	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.Set("logtostderr", "false")
	flag.Set("alsologtostderr", "false")
	pflag.Set("logtostderr", "false")
	pflag.Set("alsologtostderr", "false")
	pflag.Parse()

	if err := processFlags(); err != nil {
		klog.Exitf("processing error: %v", err)
	}
}
