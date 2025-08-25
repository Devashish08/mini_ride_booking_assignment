package events

import "driver_svc/internal/models"

type BookingCreated struct {
	BookingID  string          `json:"booking_id"`
	PickupLoc  models.Location `json:"pickuploc"`
	Dropoff    models.Location `json:"dropoff"`
	Price      int             `json:"price"`
	RideStatus string          `json:"ride_status"`
}
