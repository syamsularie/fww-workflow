package model

type ReqKtp struct {
	KTP string `json:"ktp"`
}

type ReqEmail struct {
	ReservationId int `json:"reservation_id"`
}
