package models

import "time"

type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Driver struct {
	DriverID    string `json:"driver_id"`
	Name        string `json:"name"`
	IsAvailable bool   `json:"is_available"`
}

type JobStatus string

const (
	JobStatusOpen  JobStatus = "Open"
	JobStatusTaken JobStatus = "Taken"
)

type Job struct {
	BookingID        string    `json:"booking_id"`
	PickupLoc        Location  `json:"pickuploc"`
	Dropoff          Location  `json:"dropoff"`
	Price            int       `json:"price"`
	Status           JobStatus `json:"status"`
	AcceptedDriverID *string   `json:"accepted_driver_id,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}
