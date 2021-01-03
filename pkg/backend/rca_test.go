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
			ImageURL:     "https://cali-content.usedirect.com/Images/California/ParkImages/Place/695.jpg",
			Availability: []campwiz.Availability{},
			Features:     []string{""},
		},
		{
			ResURL:   "https://www.reservecalifornia.com/",
			ResID:    "622",
			Name:     "Butano SP",
			Distance: 7,
			Desc: strings.Join([]string{
				"Butano State Park is a 3,200-acre redwood park in the Santa Cruz",
				" Mountains, featuring excellent hiking through the redwood grove",
				"s. Only three miles from the coast, the park's trails offer view",
				"s of the ocean.",
			}, ""),
			ImageURL:     "https://cali-content.usedirect.com/Images/California/ParkImages/Place/622.jpg",
			Availability: []campwiz.Availability{},
			Features:     []string{""},
		},
		{
			ResURL:   "https://www.reservecalifornia.com/",
			ResID:    "655",
			Name:     "Henry Cowell Redwoods SP",
			Distance: 13,
			Desc: strings.Join([]string{
				"Henry Cowell Redwoods State Park features 15 miles of hiking and",
				" riding trails through a forest that looks much the same as it d",
				"id 200 years ago.  Zayante Indians once lived in the area, where",
				" they found shelter, water and game.  The park is the home of th",
				"e Redwood grove, with a self-guided nature path, and Douglas fir",
				", madrone, oak and the most unusual feature of the park, a stand",
				" of Ponderosa pine. The park has a picnic area above the San Lor",
				"enzo River. Anglers fish for steelhead and salmon during the win",
				"ter. The Park has a nature center, bookstore and campfire center",
				".",
			}, ""),
			ImageURL:     "https://cali-content.usedirect.com/Images/California/ParkImages/Place/655.jpg",
			Availability: []campwiz.Availability{},
			Features:     []string{""},
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("parseResp() mismatch (-want +got):\n%s", diff)
	}
}
