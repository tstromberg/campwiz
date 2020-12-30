// The "main" package contains the HTTP server
package main

import (
	goflag "flag"
	"fmt"
	"net/http"
	"os"

	pflag "github.com/spf13/pflag"
	"k8s.io/klog/v2"

	"github.com/tstromberg/campwiz/pkg/cache"
	"github.com/tstromberg/campwiz/pkg/metadata"
	"github.com/tstromberg/campwiz/pkg/relpath"
	"github.com/tstromberg/campwiz/pkg/search"
	"github.com/tstromberg/campwiz/pkg/site"
)

var (
	persistBackendFlag           = pflag.String("persist-backend", "", "Cache persistence backend (disk, mysql, cloudsql)")
	persistPathFlag              = pflag.String("persist-path", "", "Where to persist cache to (automatic)")
	portFlag                     = pflag.Int("port", 8080, "port to run server at")
	siteFlag                     = pflag.String("site", "site/", "path to site files")
	thirdPartyFlag               = pflag.String("3p", "third_party/", "path to 3rd party files")
	providersFlag      *[]string = pflag.StringSlice("providers", search.DefaultProviders, "site providers to include")

	latFlag *float64 = pflag.Float64("lat", 37.4092297, "latitude to search from")
	lonFlag *float64 = pflag.Float64("lon", -122.07237049999999, "longitude to search from")
)

func main() {
	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	pflag.Parse()

	cs, err := cache.New(cache.Config{MaxAge: cache.RecommendedMaxAge})
	if err != nil {
		klog.Exitf("error: %w", err)
	}

	srcs, props, err := metadata.LoadAll()
	if err != nil {
		klog.Exitf("loadall failed: %v", err)
	}

	s := site.New(&site.Config{
		BaseDirectory: relpath.Find(*siteFlag),
		Cache:         cs,
		Sources:       srcs,
		Properties:    props,
		Providers:     *providersFlag,
		Latitude:      *latFlag,
		Longitude:     *lonFlag,
	})

	listenAddr := fmt.Sprintf(":%s", os.Getenv("PORT"))
	if listenAddr == ":" {
		listenAddr = fmt.Sprintf(":%d", *portFlag)
	}

	http.HandleFunc("/", s.Root())
	http.HandleFunc("/search", s.Search())
	http.HandleFunc("/healthz", s.Healthz())
	http.HandleFunc("/threadz", s.Threadz())
	klog.Infof("Listening at: %s", listenAddr)
	klog.Fatal(http.ListenAndServe(listenAddr, nil))
}
