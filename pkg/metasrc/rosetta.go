package metasrc

type RosettaEntry struct {
	ID              string `yaml:"id,omitempty"`
	ReservationName string `yaml:"reservation_name,omitempty"`
	ReservationURL  string `yaml:"reservation_url,omitempty"`
	Locale          string `yaml:"locale,omitempty"`
	URL             string `yaml:"url,omitempty"`
	GoogleMaps      string `yaml:"gmaps,omitempty"`
	Yelp            string `yaml:"yelp,omitempty"`
	TripAdvisor     string `yaml:"tripadvisor,omitempty"`
	Dyrt            string `yaml:"dyrt,omitempty"`
	CalBestCamping  string `yaml:"cal_best,omitempty"`
	Campendium      string `yaml:"campendium,omitempty"`
	Stienstra       string `yaml:"stienstra,omitempty"`
}
