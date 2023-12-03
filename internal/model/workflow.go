package model

type BookingVariables struct {
	ReservationID  int    `json:"reservationId"`
	BlacklistUser  bool   `json:"blacklistUser"`
	PeduliLindungi string `json:"peduliLindungi"`
	Dukcapil       string `json:"dukcapil"`
	StatusPayment  bool   `json:"status_payment"`
}
