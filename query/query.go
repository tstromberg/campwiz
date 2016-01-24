// The autocamper package contains all of the brains for querying campsites.
package query

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/tstromberg/autocamper/cache"
)

var (
	// raURL is the search URL to request reservation information from.
	baseURL = "http://www.reserveamerica.com/"

	// searchPageExpiry is how long search pages can be cached for.
	searchPageExpiry = time.Duration(6*3600) * time.Second

	// date format used by reserveamerica
	searchDate = "Mon Jan 2 2006"

	// amount of time to sleep between uncached fetches
	uncachedDelay = time.Millisecond * 750

	// regexp for mileage parsing
	mileageRegex = regexp.MustCompile(`(\d+\.\d+)mi`)

	// regexp for availability parsing
	availableRegex = regexp.MustCompile(`(\d+) matching`)
)

// SearchCriteria defines a list of attributes that can be sent to ReserveAmerica.
type Criteria struct {
	Lat         float64
	Lon         float64
	Date        time.Time
	Nights      int
	MaxDistance int
	MaxPages    int
}

type Result struct {
	Name          string
	ContractCode  string
	ParkId        int
	Distance      float64
	State         string
	ShortDesc     string
	MatchingSites int64
	URL           string
}

// firstPage creates the initial request object for a search.
func firstPage(c Criteria) cache.Request {
	// % curl -L -vvv 'http://www.reserveamerica.com/unifSearch.do' -H 'Content-Type: application/x-www-form-urlencoded' --data 'locationCriteria=SAN+FRANCISCO%2C+CA%2C+USA&locationPosition=%3A%3A-122.41941550000001%3A37.7749295%3A%3ACA&interest=camping&lookingFor=2003&campingDate=Sat+Jan+30+2016&lengthOfStay=2'

	v := url.Values{
		"locationCriteria": {"San Francisco, CA"},
		"locationPosition": {fmt.Sprintf("::%3.14f:%3.7f::CA", c.Lat, c.Lon)},
		"interest":         {"camping"},
		"lookingFor":       {"2003"},
		"campingDate":      {c.Date.Format(searchDate)},
		"lengthOfStay":     {strconv.Itoa(c.Nights)},
	}

	return cache.Request{
		Method:   "POST",
		URL:      baseURL + "/unifSearch.do",
		Referrer: baseURL,
		Form:     v,
		MaxAge:   searchPageExpiry,
	}
}

// nextPage creates requests for subsequent pages.
func nextPage(c Criteria, r cache.Result, page int) cache.Request {
	url := fmt.Sprintf("%s/unifSearchResults.do?currentPage=%d&paging=true&facilityType=all&agencyKey=&facilityAvailable=show_all&viewType=view_list&selectedLetter=ALL&owner=&hiddenFilters=false", baseURL, page)
	return cache.Request{
		Method:   "GET",
		URL:      url,
		Referrer: r.URL,
		Cookies:  r.Cookies,
		MaxAge:   searchPageExpiry,
	}
}

// Search performs a query against the ReserveAmerica site, returning parsed results.
func Search(crit Criteria) ([]Result, error) {
	log.Printf("Search: %+v", crit)
	r, err := cache.Fetch(firstPage(crit))
	if err != nil {
		return nil, err
	}

	parsed, err := parseResultsPage(r.Body, r.URL, 1)
	if err != nil {
		return nil, err
	}

	for i := 1; i < crit.MaxPages; i++ {
		r, err := cache.Fetch(nextPage(crit, r, i))
		if err != nil {
			return parsed, err
		}

		pr, err := parseResultsPage(r.Body, r.URL, i+1)
		if err != nil {
			return parsed, err
		}

		parsed = append(parsed, pr...)
		if !r.Cached {
			log.Printf("Previous request was uncached, sleeping ...")
			time.Sleep(uncachedDelay)
		}
	}
	return parsed, nil
}

// parseError returns a nice error message with a debug file.
func parseError(e error, body []byte) error {
	f, err := ioutil.TempFile("", "query")
	defer f.Close()
	if err != nil {
		return fmt.Errorf("parse error: %v (unable to save: %v)", e, err)
	}
	ioutil.WriteFile(f.Name(), body, 0444)
	return fmt.Errorf("parse error: %v - saved body to %s", e, f.Name())
}

// parse the results of a search page
func parseResultsPage(body []byte, sourceURL string, expectedPage int) ([]Result, error) {

	source, err := url.Parse(sourceURL)
	if err != nil {
		return nil, err
	}
	log.Printf("Parsing %s (%d bytes)", sourceURL, len(body))

	buf := bytes.NewBuffer(body)
	doc, err := goquery.NewDocumentFromReader(buf)
	if err != nil {
		return nil, parseError(err, body)
	}

	rl := doc.Find("div.usearch_results_label").First().Text()
	log.Printf("Results label: %s", rl)

	ps := doc.Find("select#pageSelector option")
	if ps.Length() == 0 {
		return nil, parseError(fmt.Errorf("Could not find select#pageSelector"), body)
	}

	page := -1
	for i := range ps.Nodes {
		opt := ps.Eq(i)
		_, exists := opt.Attr("selected")
		if exists {
			// The real value is in the "value" field.
			page, err = strconv.Atoi(opt.Text())
			if err != nil {
				return nil, parseError(err, body)
			}
		}
	}

	log.Printf("I am on page %d", page)
	if page != expectedPage {
		return nil, parseError(fmt.Errorf("page=%d, expected %d", page, expectedPage), body)
	}

	var results []Result
	sel := doc.Find("div.facility_view_card")
	for i := range sel.Nodes {
		card := sel.Eq(i)
		log.Printf("Found %d: %+v", i, card)
		r := Result{}
		link := card.Find("a.facility_link")
		r.Name = link.Text()
		href, exists := link.Attr("href")
		if !exists {
			return results, parseError(fmt.Errorf("Could not find %s href", link.Text()), body)
		} else {
			target, err := url.Parse(href)
			if err != nil {
				return nil, fmt.Errorf("Could not parse href %s: %v", href, err)
			} else {
				r.URL = source.ResolveReference(target).String()
			}
		}
		// Parse distance
		mm := mileageRegex.FindStringSubmatch(card.Find("span.sufix").Text())
		if len(mm) > 0 {
			distance, err := strconv.ParseFloat(mm[1], 64)
			if err != nil {
				return results, err
			} else {
				r.Distance = distance
			}
		}

		// Parse Matching sites
		sm := availableRegex.FindStringSubmatch(card.Find("h2").Text())
		if len(sm) > 0 {
			matching, err := strconv.ParseInt(sm[1], 10, 64)
			if err != nil {
				return results, err
			} else {
				r.MatchingSites = matching
			}
		}
		results = append(results, r)
	}

	if len(results) == 0 {
		return nil, parseError(fmt.Errorf("Unable to parse entries from body"), body)
	}

	return results, nil
}
