package repository

import (
	"context"

	"booking_svc/internal/models"
)

type CreateBookingParams struct {
	BookingID  string
	PickupLoc  models.Location
	Dropoff    models.Location
	Price      int
	RideStatus models.RideStatus
	DriverID   *string
}

type BookingRepository interface {
	Create(ctx context.Context, params CreateBookingParams) (models.Booking, error)
	ListAll(ctx context.Context) ([]models.Booking, error)
}
