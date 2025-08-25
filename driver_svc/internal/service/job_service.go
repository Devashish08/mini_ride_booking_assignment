package service

import (
	"context"
	"errors"
	"log/slog"

	"driver_svc/internal/events"
	"driver_svc/internal/models"
	"driver_svc/internal/repository"
)

var ErrJobAlreadyTaken = errors.New("job already taken")
var ErrDriverNotFound = errors.New("driver not found")

type AcceptedEventProducer interface {
	ProduceBookingAccepted(ctx context.Context, evt events.BookingAccepted) error
}

type JobsService interface {
	ListDrivers(ctx context.Context) ([]models.Driver, error)
	ListOpenJobs(ctx context.Context) ([]models.Job, error)
	AcceptJob(ctx context.Context, bookingID string, driverID string) error
}

type jobsService struct {
	drivers  repository.DriverRepository
	jobs     repository.JobRepository
	producer AcceptedEventProducer
	logger   *slog.Logger
}

func NewJobsService(dr repository.DriverRepository, jr repository.JobRepository, prod AcceptedEventProducer, logger *slog.Logger) *jobsService {
	return &jobsService{drivers: dr, jobs: jr, producer: prod, logger: logger}
}

func (s *jobsService) ListDrivers(ctx context.Context) ([]models.Driver, error) {
	return s.drivers.ListAll(ctx)
}

func (s *jobsService) ListOpenJobs(ctx context.Context) ([]models.Job, error) {
	return s.jobs.ListOpenJobs(ctx)
}

func (s *jobsService) AcceptJob(ctx context.Context, bookingID string, driverID string) error {
	d, ok, err := s.drivers.GetByID(ctx, driverID)
	if err != nil {
		return err
	}
	if !ok || !d.IsAvailable {
		return ErrDriverNotFound
	}

	won, err := s.jobs.TryAccept(ctx, bookingID, driverID)
	if err != nil {
		return err
	}
	if !won {
		return ErrJobAlreadyTaken
	}

	evt := events.BookingAccepted{
		BookingID:  bookingID,
		DriverID:   driverID,
		RideStatus: "Accepted",
	}
	return s.producer.ProduceBookingAccepted(ctx, evt)
}
