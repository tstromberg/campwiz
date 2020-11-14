// The campwiz package contains all of the brains for querying campsites.
package query

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/tstromberg/campwiz/pkg/cache"
	"github.com/tstromberg/campwiz/pkg/result"
	"k8s.io/klog/v2"
)

var (
	// raURL is the search URL to request reservation information from.
	baseURL = "https://www.reserveamerica.com"

	// searchPageExpiry is how long search pages can be cached for.
	searchPageExpiry = time.Duration(6*3600) * time.Second

	// the date format used
	campingDateFormat = "Mon Jan 2 2006"

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
	Dates       []time.Time
	Nights      int
	MaxDistance int
	MaxPages    int

	IncludeStandard bool
	IncludeGroup    bool
	IncludeBoatIn   bool
	IncludeWalkIn   bool
}

// firstPage creates the initial request object for a search.
func firstPage(c Criteria, t time.Time) cache.Request {
	// % curl -L -vvv 'http://www.reserveamerica.com/unifSearch.do' -H 'Content-Type: application/x-www-form-urlencoded' --data 'locationCriteria=SAN+FRANCISCO%2C+CA%2C+USA&locationPosition=%3A%3A-122.41941550000001%3A37.7749295%3A%3ACA&interest=camping&lookingFor=2003&campingDate=Sat+Jan+30+2016&lengthOfStay=2'

	v := url.Values{
		"locationCriteria": {"San Francisco, CA"},
		"locationPosition": {fmt.Sprintf("::%3.14f:%3.7f::CA", c.Lat, c.Lon)},
		"interest":         {"camping"},
		"lookingFor":       {"2003"},
		"campingDate":      {t.Format(campingDateFormat)},
		"lengthOfStay":     {strconv.Itoa(c.Nights)},
	}

	r := cache.Request{
		Method:   "POST",
		URL:      baseURL + "/unifSearch.do",
		Referrer: baseURL,
		Form:     v,
		MaxAge:   searchPageExpiry,
	}
	klog.Infof("First page: %s", r.URL)
	for v, k := range v {
		klog.Infof("Form value %s = %q", v, k)
	}
	return r
}

// nextPage creates requests for subsequent pages.
func nextPage(r cache.Result, page int) cache.Request {
	url := fmt.Sprintf("%s/unifSearchResults.do?currentPage=%d&paging=true&facilityType=all&agencyKey=&facilityAvailable=show_all&viewType=view_list&selectedLetter=ALL&activityType=&owner=&hiddenFilters=false", baseURL, page)
	return cache.Request{
		Method:   "GET",
		URL:      url,
		Referrer: r.URL,
		Cookies:  r.Cookies,
		MaxAge:   searchPageExpiry,
	}
}

// SearchDate runs a search for a single date
func SearchDate(crit Criteria, date time.Time) (result.Results, error) {
	klog.Infof("searchForDate: %+v", crit)

	// This page is going to redirect you.
	r, err := cache.Fetch(firstPage(crit, date))
	if err != nil {
		return nil, err
	}

	parsed, err := parseResultsPage(r.Body, r.URL, date, 1)
	if err != nil {
		return nil, err
	}

	for i := 1; i < crit.MaxPages; i++ {
		r, err := cache.Fetch(nextPage(r, i))
		if err != nil {
			return parsed, err
		}

		pr, err := parseResultsPage(r.Body, r.URL, date, i+1)
		if err != nil {
			return parsed, err
		}

		parsed = append(parsed, pr...)
		if !r.Cached {
			klog.V(1).Infof("Previous request was uncached, sleeping ...")
			time.Sleep(uncachedDelay)
		}
	}
	return parsed, nil
}

// Search performs a RA, returns parsed results.
func Search(crit Criteria) (result.Results, error) {
	var results result.Results
	for _, d := range crit.Dates {
		dr, err := searchForDate(crit, d)
		if err != nil {
			return results, err
		}
		results = append(results, dr...)
	}
	klog.Infof("Found %d results", len(results))
	filtered := filter(crit, results)
	klog.Infof("Post-filter: %d results", len(filtered))
	merged := merge(filtered)
	klog.Infof("Post-merge: %d results", len(merged))
	return merged, nil
}

// parseError returns a nice error message with a debug file.
func parseError(e error, body []byte) error {
	f, err := ioutil.TempFile("", "query.*.html")
	defer f.Close()
	if err != nil {
		return fmt.Errorf("parse error: %v (unable to save: %v)", e, err)
	}
	err = ioutil.WriteFile(f.Name(), body, 0444)
	if err != nil {
		return err
	}
	return fmt.Errorf("parse error: %v - saved body to %s", e, f.Name())
}

// availableSiteCounts returns the number of single & group sites available for a card.
func availableSiteCounts(card *goquery.Selection, amenities string) (result.Availability, error) {
	a := result.Availability{}

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
				klog.V(1).Infof("Day: %s (%d)", ctype, count)
				continue
			}
			if strings.Contains(ctype, "GROUP") {
				a.Group += count
				klog.V(1).Infof("Group: %s (%d)", ctype, count)
				continue
			}
			if strings.Contains(ctype, "RV/TRAILER") || strings.Contains(ctype, "RV ELECTRIC") {
				a.Rv += count
				klog.V(1).Infof("Rv: %s (%d)", ctype, count)
				continue
			}
			if strings.Contains(ctype, "HORSE") || strings.Contains(ctype, "EQUESTRIAN") {
				a.Equestrian += count
				klog.V(1).Infof("Equestrian: %s (%d)", ctype, count)
				continue
			}
			if strings.Contains(ctype, "WALK") || strings.Contains(ctype, "HIKE") {
				a.WalkIn += count
				klog.V(1).Infof("WalkIn: %s (%d)", ctype, count)
				continue
			}
			if strings.Contains(ctype, "BOAT") || strings.Contains(ctype, "FLOAT") {
				a.Boat += count
				klog.V(1).Infof("Boat: %s (%d)", ctype, count)
				continue
			}

			// We have no way of knowing how many sites are accessible or not :(
			if strings.Contains(amenities, "Accessible") && a.Accessible == 0 {
				klog.V(1).Infof("Accessible: %s (%d)", ctype, 1)
				a.Accessible = 1
				count = count - 1
			}

			if count > 0 {
				klog.V(1).Infof("Standard: %s (%d)", ctype, count)
				a.Standard += count
			}
		}
	}
	return a, nil
}

func parseCard(source *url.URL, card *goquery.Selection, date time.Time) (result.Result, error) {
	klog.Infof("Parsing card: %s", card.Text())
	r := result.Result{}
	link := card.Find("a.facility_link")
	r.Name = link.Text()

	r.ShortDesc = strings.Replace(card.Find("span.description").First().Text(), "[more]", "", 1)
	href, exists := link.Attr("href")
	if !exists {
		return r, fmt.Errorf("Could not find %s href", link.Text())
	}
	klog.Infof("Site URL: %s", href)

	target, err := url.Parse(href)
	if err != nil {
		return r, fmt.Errorf("Could not parse href %s: %v", href, err)
	}
	r.URL = source.ResolveReference(target).String()

	pids := target.Query()["parkId"]
	if len(pids) == 0 {
		return r, fmt.Errorf("%s has no parkId", href)
	}
	r.ParkId = pids[0]
	r.ContractCode = target.Query()["contractCode"][0]
	// Parse distance
	mm := mileageRegex.FindStringSubmatch(card.Find("span.sufix").Text())
	if len(mm) > 0 {
		distance, err := strconv.ParseFloat(mm[1], 64)
		if err != nil {
			return r, err
		}
		r.Distance = distance
	}

	// Parse amenities
	r.Amenities = card.Find("div.sites_amenities").First().Text()
	klog.V(1).Infof("Amenities: %s", r.Amenities)

	// Parse Matching sites
	a, err := availableSiteCounts(card, r.Amenities)
	a.Date = date
	if err != nil {
		return r, err
	}
	r.Availability = append(r.Availability, a)
	klog.Infof("Card result: %+v", r)
	return r, nil
}

// parse the results of a search page
func parseResultsPage(body []byte, sourceURL string, date time.Time, expectedPage int) (result.Results, error) {
	klog.Infof("*************** Parsing %s - expected page: %d", sourceURL, expectedPage)
	source, err := url.Parse(sourceURL)
	if err != nil {
		return nil, err
	}
	klog.V(1).Infof("Parsing %s (%d bytes)", sourceURL, len(body))

	buf := bytes.NewBuffer(body)
	doc, err := goquery.NewDocumentFromReader(buf)
	if err != nil {
		return nil, parseError(err, body)
	}

	rl := doc.Find("div.facility_view_header_near").First().Text()
	klog.V(1).Infof("Results label: %q", rl)

	// Find the marker that tells us what page we are on.
	ps := doc.Find("select[name=pageSelector] option")
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

	klog.V(1).Infof("I am on page %d", page)
	if page != expectedPage {
		return nil, parseError(fmt.Errorf("page=%d, expected %d", page, expectedPage), body)
	}

	var results result.Results
	sel := doc.Find("div.facility_view_card")
	for i := range sel.Nodes {
		r, err := parseCard(source, sel.Eq(i), date)
		if err != nil {
			klog.Warningf("Unable to parse card %d: %v", i, err)
			continue
		}
		results = append(results, r)
	}

	if len(results) == 0 {
		return nil, parseError(fmt.Errorf("Unable to parse entries from body"), body)
	}

	klog.Infof("Finished parsing %s - %d results", sourceURL, len(results))
	return results, nil
}
