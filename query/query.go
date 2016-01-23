// The autocamper package contains all of the brains for querying campsites.
package query

import (
	"bytes"
	"log"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/tstromberg/autocamper/cache"
)

var (
	// reserveAmericaUrl is the search URL to request reservation information from.
	reserveAmericaUrl = "http://www.reserveamerica.com/unifSearch.do"

	// searchPageExpiry is how long search pages can be cached for.
	searchPageExpiry = time.Duration(900) * time.Second
)

// SearchCriteria defines a list of attributes that can be sent to ReserveAmerica.
type SearchCriteria struct {
}

// Search performs a query against the ReserveAmerica site.
func Search(locationCriteria string) (cache.Result, error) {

	// TODO(tstromberg): Stop hardcoding values.
	v := url.Values{
		"locationCriteria":  {locationCriteria},
		"locationPosition":  {"::-122.4750292:37.7597481:"},
		"interest":          {"camping"},
		"lookingFor":        {"2003"},
		"campingDate":       {"Fri Aug 29 2016"},
		"lengthOfStay":      {"2"},
		"camping_2003_3012": {"3"},
	}

	log.Printf("POSTing to %s - values: %s", reserveAmericaUrl, v)
	res, err := cache.Fetch(cache.Request{"GET", "http://www.reserveamerica.com/unifSearch.do", v, searchPageExpiry})
	if err != nil {
		log.Printf("Err: %v", err)
	} else {
		log.Printf("Response: %d %+v", res.StatusCode, res.Header)
	}
	return res, err
}

// Parse parses the results of a ReserveAmerica search page.
func Parse(body []byte) error {
	log.Printf("Body: %s", body)

	buf := bytes.NewBuffer(body)
	doc, err := goquery.NewDocumentFromReader(buf)
	if err != nil {
		return err
	}
	log.Printf("Doc: %s", doc)

	doc.Find("a.facility_link").Each(func(i int, s *goquery.Selection) {
		log.Printf("Found %d: %s", i, s.Text())
	})

	return nil
}
