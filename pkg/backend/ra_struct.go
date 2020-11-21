package backend

type raControl struct {
	CurrentPage int
	PageSize    int
}

type raRecord struct {
	NamingID  string
	Name      string
	Proximity float64
	Details   raDetails
}

type raAvailability struct {
	Available      bool
	ReservableType string
}

type raDetails struct {
	BaseURL      string
	Availability raAvailability
}

type raResponse struct {
	TotalRecords int
	TotalPages   int
	StartIndex   int
	EndIndex     int
	Control      raControl
	Records      []raRecord
}
