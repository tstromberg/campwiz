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
	"strings"
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
	availableRegex = regexp.MustCompile(`(.*?)\((\d+)\)`)
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

type Availability struct {
	Group         int64
	Standard      int64
	Accessible    int64
	Equestrian    int64
	Day           int64
	RvTrailerOnly int64
}

type Result struct {
	Name         string
	ContractCode string
	ParkId       int64
	Distance     float64
	State        string
	ShortDesc    string
	Availability Availability
	URL          string
	Amenities    string
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

// availableSiteCounts returns the number of single & group sites available for a card.
func availableSiteCounts(card *goquery.Selection, amenities string) (Availability, error) {
	a := Availability{}

	sel := card.Find("span.site_type_item a")
	for i := range sel.Nodes {
		sm := availableRegex.FindStringSubmatch(sel.Eq(i).Text())
		if len(sm) > 0 {
			ctype := sm[1]
			count, err := strconv.ParseInt(sm[2], 10, 64)
			if err != nil {
				return a, err
			}
			if strings.Contains(ctype, "DAY") {
				a.Day += count
				log.Printf("Day: %s (%d)", ctype, count)
				continue
			}
			if strings.Contains(ctype, "GROUP") {
				a.Group += count
				log.Printf("Group: %s (%d)", ctype, count)
				continue
			}
			if strings.Contains(ctype, "RV/TRAILER") {
				a.RvTrailerOnly += count
				log.Printf("Rv/Trailer: %s (%d)", ctype, count)
				continue
			}
			if strings.Contains(ctype, "HORSE") || strings.Contains(ctype, "EQUESTRIAN") {
				a.Equestrian += count
				log.Printf("Equestrian: %s (%d)", ctype, count)
				continue
			}

			// We have no way of knowing how many sites are accessible or not :(
			if strings.Contains(amenities, "Accessible") && a.Accessible == 0 {
				log.Printf("Accessible: %s (%d)", ctype, 1)
				a.Accessible = 1
				count = count - 1
			}

			if count > 0 {
				log.Printf("Standard: %s (%d)", ctype, count)
				a.Standard += count
			}
		}
	}
	return a, nil
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
		r := Result{}
		link := card.Find("a.facility_link")
		r.Name = link.Text()
		log.Printf("Parsing: %s", r.Name)
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

		// Parse amenities
		r.Amenities = card.Find("div.sites_amenities").First().Text()
		log.Printf("Amenities: %s", r.Amenities)

		// Parse Matching sites
		a, err := availableSiteCounts(card, r.Amenities)
		if err != nil {
			return results, err
		}
		r.Availability = a
		results = append(results, r)
	}

	if len(results) == 0 {
		return nil, parseError(fmt.Errorf("Unable to parse entries from body"), body)
	}

	return results, nil
}
