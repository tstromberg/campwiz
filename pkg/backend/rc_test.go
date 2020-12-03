package backend

import (
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/tstromberg/campwiz/pkg/cache"
	"github.com/tstromberg/campwiz/pkg/campwiz"
)

func TestRCaliforniaReq(t *testing.T) {
	rc := &RCalifornia{}

	date, err := time.Parse("2006-01-02", "2021-02-12")
	if err != nil {
		t.Fatalf("time parse: %v", err)
	}
	q := campwiz.Query{
		StayLength:  4,
		Lon:         -122.07237049999999,
		Lat:         37.4092297,
		MaxDistance: 100,
	}

	got, err := rc.req(q, date)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	t.Logf("body: %s", got.Body)

	want := cache.Request{
		Method:      "POST",
		URL:         "https://calirdr.usedirect.com/rdr/rdr/search/place",
		Referrer:    "https://www.reservecalifornia.com/",
		MaxAge:      time.Duration(6 * time.Hour),
		ContentType: "application/json",
		Body:        []byte(`{"PlaceId":0,"Latitude":"37.4092","Longitude":"-122.0724","HighlightedPlaceId":0,"StartDate":"02-12-2021","Nights":"4","CountNearby":true,"NearbyLimit":100,"NearbyOnlyAvailable":true,"NearbyCountLimit":100,"Sort":"Distance","CustomerID":"0","RefreshFavourites":true,"IsADA":false,"UnitCategoryId":0,"SleepingUnitId":0,"MinVehicleLength":0,"UnitTypeGroupIds":null}`),
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("rcPageRequest() mismatch (-want +got):\n%s", diff)
	}
}

func TestRCaliforniaParse(t *testing.T) {
	ra := &RCalifornia{}

	bs, err := ioutil.ReadFile("testdata/rc_place.json")
	if err != nil {
		t.Fatalf("readfile: %v", err)
	}
	date, err := time.Parse("2006-01-02", "2021-02-12")
	if err != nil {
		t.Fatalf("time parse: %v", err)
	}
	q := campwiz.Query{
		StayLength:  4,
		Lon:         -122.07237049999999,
		Lat:         37.4092297,
		MaxDistance: 100,
	}

	got, err := ra.parse(bs, date, q)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	want := []campwiz.Result{
		{
			ResID:    "/rc/682",
			Name:     "Mount Tamalpais SP",
			Distance: 17,
			Desc: strings.Join([]string{
				"Just north of San Francisco's Golden Gate is Mount Tamalpais, 6,",
				"300 acres of redwood groves and oak woodlands with a spectacular",
				" view from the 2,571-foot peak. On a clear day, visitors can see",
				" the Farallon Islands 25 miles out to sea, the Marin County Hill",
				"s, San Francisco and the Bay, hills and cities of the East Bay, ",
				"and Mount Diablo. On rare occasions, the Sierra Nevada's snow-co",
				"vered mountains can be seen 150 miles away. Coastal Miwok Indian",
				"s lived in the area for thousands of years before Europeans arri",
				"ved. In 1770, two explorers named the mountain La Sierra de Nues",
				"tro Padre de San Francisco, which was later changed to the Miwok",
				" word Tamalpais. With the Gold Rush of 1849, San Francisco grew;",
				" and more people began to use Mount Tamalpais for recreation. Tr",
				"ails were developed, and a wagon road was built. Later, a railwa",
				"y was completed and became known as the Crookedest Railroad in t",
				"he World It was abandoned in 1930 after a wildfire damaged the l",
				"ine."}, ""),
			URL: "http://www.parks.ca.gov/?page_id=471",
			Availability: []campwiz.Availability{
				{
					SiteType: "campsite",
					Date:     time.Date(2021, 02, 12, 0, 0, 0, 0, time.UTC),
					URL:      "https://www.reservecalifornia.com/CaliforniaWebHome/Facilities/SearchViewUnitAvailabity.aspx",
				},
			},
			Features: []string{"Bicycling", "Birdwatching", "Body surfing", "Camping", "Fishing", "Group Camping", "Hiking", "Horseback riding",
				"Lodging",
				"Picnic area",
				"Scuba diving",
				"Surfing",
				"Swimming",
				"Visitor Center",
			},
		},
		{
			ResID:    "/rc/683",
			Name:     "Mount Diablo SP",
			Distance: 26,
			Desc: strings.Join([]string{
				"On a clear day, from the summit of Mount Diablo State Park visit",
				"ors can see 35 of California's 58 counties. It is said that the ",
				"view is surpassed only by that of 19,000-foot Mount Kilimanjaro ",
				"in Africa. With binoculars, Yosemite's Half Dome is even visible",
				" from Mt. Diablo. The park features exce hiking and rock climbin",
				"g opportunities. The mountain was formed when a mass of underlyi",
				"ng rock was gradually forced up through the earth's surface so, ",
				"unlike other mountains, older and older rocks are encountered as",
				" you climb the mountain. The mountain was regarded as sacred to ",
				"Native Americans.",
			}, ""),
			URL: "http://www.parks.ca.gov/?page_id=517",
			Availability: []campwiz.Availability{
				{
					SiteType: "campsite",
					Date:     time.Date(2021, 02, 12, 0, 0, 0, 0, time.UTC),
					URL:      "https://www.reservecalifornia.com/CaliforniaWebHome/Facilities/SearchViewUnitAvailabity.aspx",
				},
			},
			Features: []string{"Bicycling", "Birdwatching", "Camping", "Group Camping", "Hiking", "Horseback riding", "Museum", "Picnic area", "Visitor Center"},
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("parseResp() mismatch (-want +got):\n%s", diff)
	}
}
