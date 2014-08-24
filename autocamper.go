// autocamper is a prototype for automatically booking campsites on ReserveAmerica.
package main

import (
	//"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

var (
	// reserveAmericaUrl is the search URL to request reservation information from.
	reserveAmericaUrl = "http://www.reserveamerica.com/unifSearch.do"
)

// SearchCriteria defines a list of attributes that can be sent to ReserveAmerica.
type SearchCriteria struct {
}

// Search performs a query against the ReserveAmerica site.
func Search(locationCriteria string) (*http.Response, error) {
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
	resp, err := http.PostForm("http://www.reserveamerica.com/unifSearch.do", v)
	log.Printf("Response: %s", resp)
	if err != nil {
		log.Printf("Err: %s", err)
	}
	return resp, err
}

// Parse parses the results of a ReserveAmerica search page.
func Parse(resp *http.Response) error {
	// body, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	//	return err
	//}
	// Just debugging things here.
	//log.Printf("Body: %s", body)
	//return nil

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return err
	}
	log.Printf("Doc: %s", doc)

	doc.Find("a.facility_link").Each(func(i int, s *goquery.Selection) {
		log.Printf("Found %d: %s", i, s.Text())
	})

	return nil
}

func main() {
	resp, err := Search("94122")
	if err != nil {
		log.Fatalf("Fetch error: %s", err)
	}
	err = Parse(resp)
	if err != nil {
		log.Fatalf("Parse error: %s", err)
	}
}
