// The autocamper package contains all of the brains for querying campsites.
package query

import (
	"bytes"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/tstromberg/autocamper/cache"
)

var (
	// reserveAmericaUrl is the search URL to request reservation information from.
	reserveAmericaUrl = "http://www.reserveamerica.com/unifSearch.do"

	// searchPageExpiry is how long search pages can be cached for.
	searchPageExpiry = time.Duration(6 * 3600) * time.Second

	// date format used by reserveamerica
	searchDate = "Mon Jan 2 2006"
)

// SearchCriteria defines a list of attributes that can be sent to ReserveAmerica.
type Criteria struct {
	Lat float64
	Lon float64
	Date time.Time
	Nights	int
	MaxDistance int
	MaxPages int
}

type Result struct {
	Name string
	ContractCode string
	ParkId	int
	Distance	float64
	State string
	ShortDesc string
	MatchingSites	int
	SourceURL	string
}


// initialRequest creates the initial request object for a search.
func initialRequest(c Criteria) cache.Request {
	// % curl -L -vvv 'http://www.reserveamerica.com/unifSearch.do' -H 'Content-Type: application/x-www-form-urlencoded' --data 'locationCriteria=SAN+FRANCISCO%2C+CA%2C+USA&locationPosition=%3A%3A-122.41941550000001%3A37.7749295%3A%3ACA&interest=camping&lookingFor=2003&campingDate=Sat+Jan+30+2016&lengthOfStay=2'

	v := url.Values{
		"locationCriteria":  {"San Francisco, CA"},
		"locationPosition":  {fmt.Sprintf("::%3.14f:%3.7f::CA", c.Lat, c.Lon)},
		"interest":          {"camping"},
		"lookingFor":        {"2003"},
		"campingDate":       {c.Date.Format(searchDate)},
		"lengthOfStay":      {strconv.Itoa(c.Nights)},
	}

	url := "http://www.reserveamerica.com/unifSearch.do"
	return cache.Request{
		Method: "POST",
		URL: url,
		Referrer: "http://www.reserveamerica.com/",
		Form: v,
		MaxAge: searchPageExpiry,
	}
}

// pagingRequest creates requests for subsequent pages.
func pagingRequest(c Criteria, r cache.Result, page int) cache.Request {
	url := fmt.Sprintf("http://www.reserveamerica.com/unifSearchResults.do?currentPage=%d&paging=true&facilityType=all&agencyKey=&facilityAvailable=show_all&viewType=view_list&selectedLetter=ALL&owner=&hiddenFilters=false", page)
	return cache.Request{
		Method: "GET",
		URL: url,
		Referrer: r.URL,
		Cookies: r.Cookies,
		MaxAge: searchPageExpiry,
	}
}


// Search performs a query against the ReserveAmerica site, returning parsed results.
func Search(crit Criteria) ([]Result, error) {
	log.Printf("Search: %+v", crit)
	r, err := cache.Fetch(initialRequest(crit))
	if err != nil {
		return nil, err
	}

	parsed, err := Parse(r.Body)
	if err != nil {
		return nil, err
	}

	for i :=1; i<crit.MaxPages; i++ {
		r, err := cache.Fetch(pagingRequest(crit, r, i))
		if err != nil {
			return parsed, err
		}

		pr, err := Parse(r.Body)
		if err != nil {
			return parsed, err
		}

		parsed = append(parsed, pr...)
		if ! r.Cached {
			log.Printf("Previous request was uncached, sleeping ...")
			time.Sleep(500 * time.Millisecond)
		}
	}
	return parsed, nil
}

// Parse parses the results of a ReserveAmerica search page.
func Parse(body []byte) ([]Result, error) {
	var results []Result
	log.Printf("Body: %s", body)

	buf := bytes.NewBuffer(body)
	doc, err := goquery.NewDocumentFromReader(buf)
	if err != nil {
		return results, err
	}
	log.Printf("Doc: %s", doc)


	doc.Find("a.facility_link").Each(func(i int, s *goquery.Selection) {
		log.Printf("Found %d: %s", i, s.Text())
		r := Result{Name: s.Text()}
		results = append(results, r)
	})

	return results, nil
}
