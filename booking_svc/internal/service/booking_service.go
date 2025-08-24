package service

import (
	"context"
	"log/slog"

	"booking_svc/internal/events"
	"booking_svc/internal/models"
	"booking_svc/internal/mq"
	"booking_svc/internal/repository"

	"github.com/google/uuid"
)

type CreateBookingInput struct {
	PickupLoc models.Location
	Dropoff   models.Location
	Price     int
}

type BookingService interface {
	CreateBooking(ctx context.Context, in CreateBookingInput) (models.Booking, error)
	ListBookings(ctx context.Context) ([]models.Booking, error)
}

type bookingService struct {
	repo     repository.BookingRepository
	producer *mq.Producer
	logger   *slog.Logger
}

func NewBookingService(repo repository.BookingRepository, producer *mq.Producer, logger *slog.Logger) BookingService {
	return &bookingService{repo: repo, producer: producer, logger: logger}
}

func (s *bookingService) CreateBooking(ctx context.Context, in CreateBookingInput) (models.Booking, error) {
	bookingID := uuid.NewString()
	rideStatus := models.RideStatusRequested
	var driverID *string

	created, err := s.repo.Create(ctx, repository.CreateBookingParams{
		BookingID:  bookingID,
		PickupLoc:  in.PickupLoc,
		Dropoff:    in.Dropoff,
		Price:      in.Price,
		RideStatus: rideStatus,
		DriverID:   driverID,
	})
	if err != nil {
		return models.Booking{}, err
	}

	evt := events.BookingCreated{
		BookingID:  created.BookingID,
		PickupLoc:  created.PickupLoc,
		Dropoff:    created.Dropoff,
		Price:      created.Price,
		RideStatus: string(created.RideStatus),
	}
	if err := s.producer.ProduceBookingCreated(ctx, evt); err != nil {
		// Strong consistency for assignment: fail request if event not produced
		return models.Booking{}, err
	}
	return created, nil
}

func (s *bookingService) ListBookings(ctx context.Context) ([]models.Booking, error) {
	return s.repo.ListAll(ctx)
}
