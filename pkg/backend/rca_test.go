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
		Body:        []byte(`{"googlePlaceSearchParameters":{"Latitude":"37.17159","Longitude":"-122.22203","South":37.00781829886819,"North":37.335007514028106,"East":-121.96076138427298,"West":-122.48329861572044,"Filter":true,"BackToHome","ZoomLevel":9,"CenterLatitude":37.17159,"CenterLongitude":-122.22203,"ChangeDragandZoom":true,"BacktoFacility":true,"ChooseActivity":null,"IsFilterClick":false,"AvailabilitySearchParams":{"RegionId":0,"PlaceId":[","FacilityId":0,"StartDate":"01/04/2021","Nights":"1","CategoryId":0,"UnitTypeIds","UnitTypesCategory","ShowOnlyAdaUnits":false,"ShowOnlyTentSiteUnits":"false","ShowOnlyRvSiteUnits":"false","MinimumVehicleLength":"0","PageIndex":0,"PageSize","Page1","NoOfRecords":100,"ShowSiteUnitsName":"0","Autocomplitename":"Big Basin Redwoods SP","ParkFinder","ParkCategory":8,"ChooseActivity":"1","IsPremium":false},"IsFacilityLevel":false,"PlaceIdFacilityLevel":0,"MapboxPlaceid","Screenresolution":1421}'`),
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("rcPageRequest() mismatch (-want +got):\n%s", diff)
	}
}

func TestRCaliforniaAdvParse(t *testing.T) {
	ra := &RCaliforniaAdv{}

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
			ResURL:   "https://www.reservecalifornia.com/",
			ResID:    "695",
			Name:     "Portola Redwoods SP",
			Distance: 6,
			Desc: strings.Join([]string{
				"Portola Redwoods State Park has a rugged, natural basin forested",
				" with coast redwoods, Douglas fir and live oak. Eighteen miles o",
				"f trails crisscross the canyon and its two streams, Peters Creek",
				" and Pescadero Creek. A short nature trail along Pescadero Creek",
				" introduces visitors to the natural history of the area. Visitor",
				"s can see clam shells and other marine deposits from the time wh",
				"en the area was once covered by the ocean. The park has one of t",
				"he tallest redwoods (300 feet high) in the Santa Cruz Mountains.",
			}, ""),
			URL:      "http://www.parks.ca.gov/?page_id=539",
			ImageURL: "https://cali-content.usedirect.com/Images/California/ParkImages/Place/695.jpg",
			Availability: []campwiz.Availability{
				{
					Kind:      campwiz.Tent,
					Name:      "Portola Campground",
					Desc:      "Tent Campsite",
					SpotCount: 12,
					Date:      time.Date(2021, 0o2, 12, 0, 0, 0, 0, time.UTC),
				},
				{
					Kind:      campwiz.Standard,
					Name:      "Portola Campground",
					Desc:      "Campsite",
					SpotCount: 2,
					Date:      time.Date(2021, 0o2, 12, 0, 0, 0, 0, time.UTC),
				},
				{
					Kind:      campwiz.Walk,
					Name:      "Family Walk in",
					Desc:      "Hike in Campsite",
					SpotCount: 4,
					Date:      time.Date(2021, 0o2, 12, 0, 0, 0, 0, time.UTC),
				},
				{
					Kind:      campwiz.Group,
					Name:      "Redwoods Group Camping Area",
					Desc:      "Group Campsite",
					SpotCount: 3,
					Date:      time.Date(2021, 0o2, 12, 0, 0, 0, 0, time.UTC),
				},
			},
			Features: []string{"Bicycling", "Camping", "Group Camping", "Hiking", "Museum", "Picnic area", "Swimming", "Visitor Center"},
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("parseResp() mismatch (-want +got):\n%s", diff)
	}
}
