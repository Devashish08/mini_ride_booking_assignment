package models

import "time"

type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type RideStatus string

const (
	RideStatusRequested RideStatus = "Requested"
	RideStatusAccepted  RideStatus = "Accepted"
)

type Booking struct {
	BookingID  string     `json:"booking_id"`
	PickupLoc  Location   `json:"pickuploc"`
	Dropoff    Location   `json:"dropoff"`
	Price      int        `json:"price"`
	RideStatus RideStatus `json:"ride_status"`
	DriverID   *string    `json:"driver_id,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}
