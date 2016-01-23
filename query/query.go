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
	// % curl -L -vvv 'http://www.reserveamerica.com/unifSearch.do' -H 'Content-Type: application/x-www-form-urlencoded' --data 'locationCriteria=SAN+FRANCISCO%2C+CA%2C+USA&locationPosition=%3A%3A-122.41941550000001%3A37.7749295%3A%3ACA&interest=camping&lookingFor=2003&campingDate=Sat+Jan+30+2016&lengthOfStay=2'

	// TODO(tstromberg): Stop hardcoding values.
	v := url.Values{
		"locationCriteria":  {locationCriteria},
		"locationPosition":  {"::-122.41941550000001:37.7749295::CA"},
		"interest":          {"camping"},
		"lookingFor":        {"2003"},
		"campingDate":       {"Sat Jan 30 2016"},
		"lengthOfStay":      {"2"},
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
