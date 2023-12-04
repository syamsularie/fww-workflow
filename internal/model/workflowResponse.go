package model

type RegulationtResponse struct {
	Status bool `json:"status"`
}

type RegulationtResponseString struct {
	Status string `json:"status"`
}

type PaymentDetailResponse struct {
	FlightNumber       string  `json:"flight_number"`
	PassengerFirstName string  `json:"passenger_first_name"`
	PassengerLastName  string  `json:"passenger_last_name"`
	SeatNumber         int     `json:"seat_number"`
	Price              float32 `json:"price"`
	PaymentStatus      bool    `json:"payment_status"`
	PaymentMethod      string  `json:"payment_method"`
	PaymentCode        string  `json:"payment_code"`
}
