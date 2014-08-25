// The autocamper package contains all of the brains for querying campsites.
package autocamper

import (
	"bytes"
	"log"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	// reserveAmericaUrl is the search URL to request reservation information from.
	reserveAmericaUrl = "http://www.reserveamerica.com/unifSearch.do"

	// searchPageExpiry is how long search pages can be cached for.
	searchPageExpiry = time.Duration(86400) * time.Second
)

// SearchCriteria defines a list of attributes that can be sent to ReserveAmerica.
type SearchCriteria struct {
}

// Search performs a query against the ReserveAmerica site.
func Search(locationCriteria string) (CachedHttpResponse, error) {

	// TODO(tstromberg): Stop hardcoding values.
	v := url.Values{
		"locationCriteria":  {locationCriteria},
		"locationPosition":  {"::-122.4750292:37.7597481:"},
		"interest":          {"camping"},
		"lookingFor":        {"2003"},
		"campingDate":       {"Fri Aug 29 2014"},
		"lengthOfStay":      {"2"},
		"camping_2003_3012": {"3"},
	}

	log.Printf("POSTing to %s - values: %s", reserveAmericaUrl, v)
	resp, err := cachedFetch("http://www.reserveamerica.com/unifSearch.do", v, searchPageExpiry)
	log.Printf("Response: %s", resp.StatusCode)
	if err != nil {
		log.Printf("Err: %s", err)
	}
	return resp, err
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
