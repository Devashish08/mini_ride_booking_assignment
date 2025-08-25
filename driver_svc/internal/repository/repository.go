package repository

import (
	"context"

	"driver_svc/internal/models"
)

type DriverRepository interface {
	ListAll(ctx context.Context) ([]models.Driver, error)
	GetByID(ctx context.Context, driverID string) (models.Driver, bool, error)
}

type UpsertJobParams struct {
	BookingID string
	PickupLoc models.Location
	Dropoff   models.Location
	Price     int
}

type JobRepository interface {
	UpsertOpenJob(ctx context.Context, p UpsertJobParams) error
	ListOpenJobs(ctx context.Context) ([]models.Job, error)
	TryAccept(ctx context.Context, bookingID string, driverID string) (bool, error)
}
