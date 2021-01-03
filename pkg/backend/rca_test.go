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

func TestRCaliforniaAdvReq(t *testing.T) {
	rc := &RCaliforniaAdv{}

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
		URL:         "https://www.reservecalifornia.com/CaliforniaWebHome/Facilities/AdvanceSearch.aspx/GetPlaceData",
		Referrer:    "https://www.reservecalifornia.com/CaliforniaWebHome/Facilities/AdvanceSearch.aspx",
		MaxAge:      time.Duration(6 * time.Hour),
		ContentType: "application/json",
		Body:        []byte(`{"googlePlaceSearchParameters":{"Latitude":"37.17159","Longitude":"-122.22203","South":37.00781829886819,"North":37.335007514028106,"East":-121.96076138427298,"West":-122.48329861572044,"Filter":true,"BackToHome":"","ZoomLevel":9,"CenterLatitude":37.17159,"CenterLongitude":-122.22203,"ChangeDragandZoom":true,"BacktoFacility":true,"ChooseActivity":null,"IsFilterClick":false,"AvailabilitySearchParams":{"RegionId":0,"PlaceId":["0"],"FacilityId":0,"StartDate":"01/04/2021","Nights":"1","CategoryId":0,"UnitTypeIds":[],"UnitTypesCategory":[],"ShowOnlyAdaUnits":false,"ShowOnlyTentSiteUnits":"false","ShowOnlyRvSiteUnits":"false","MinimumVehicleLength":"0","PageIndex":0,"PageSize":20,"Page1":20,"NoOfRecords":100,"ShowSiteUnitsName":"0","Autocomplitename":"Big Basin Redwoods SP","ParkFinder":[],"ParkCategory":8,"ChooseActivity":"1","IsPremium":false},"IsFacilityLevel":false,"PlaceIdFacilityLevel":0,"MapboxPlaceid":0},"Screenresolution":1421}'`),
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("rcPageRequest() mismatch (-want +got):\n%s", diff)
	}
}

func TestRCaliforniaAdvParse(t *testing.T) {
	ra := &RCalifornia{}

	bs, err := ioutil.ReadFile("testdata/rca.json")
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
				"ine.",
			}, ""),
			URL: "http://www.parks.ca.gov/?page_id=471",
			Availability: []campwiz.Availability{
				{
					SiteKind: campwiz.Tent,
					Date:     time.Date(2021, 0o2, 12, 0, 0, 0, 0, time.UTC),
					URL:      "https://www.reservecalifornia.com/CaliforniaWebHome/Facilities/SearchViewUnitAvailabity.aspx",
				},
			},
			Features: []string{
				"Bicycling", "Birdwatching", "Body surfing", "Camping", "Fishing", "Group Camping", "Hiking", "Horseback riding",
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
					SiteKind: campwiz.Tent,
					Date:     time.Date(2021, 0o2, 12, 0, 0, 0, 0, time.UTC),
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
