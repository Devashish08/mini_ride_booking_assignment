package handlerhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"booking_svc/internal/models"
	"booking_svc/internal/service"

	"github.com/go-chi/chi/v5"
)

type fakeBookingService struct {
	createFn func(ctx context.Context, in service.CreateBookingInput) (models.Booking, error)
	listFn   func(ctx context.Context) ([]models.Booking, error)
}

func (f *fakeBookingService) CreateBooking(ctx context.Context, in service.CreateBookingInput) (models.Booking, error) {
	return f.createFn(ctx, in)
}
func (f *fakeBookingService) ListBookings(ctx context.Context) ([]models.Booking, error) {
	return f.listFn(ctx)
}

func TestCreateBooking_Handler(t *testing.T) {
	now := time.Now().UTC()
	want := models.Booking{
		BookingID:  "b-1",
		PickupLoc:  models.Location{Lat: 12.9, Lng: 77.6},
		Dropoff:    models.Location{Lat: 12.95, Lng: 77.64},
		Price:      220,
		RideStatus: models.RideStatusRequested,
		CreatedAt:  now,
	}
	h := NewBookingHandler(&fakeBookingService{
		createFn: func(ctx context.Context, in service.CreateBookingInput) (models.Booking, error) { return want, nil },
		listFn:   func(ctx context.Context) ([]models.Booking, error) { return nil, nil },
	})

	r := chi.NewRouter()
	h.RegisterRoutes(r)

	t.Run("201", func(t *testing.T) {
		body := `{"pickuploc":{"lat":12.9,"lng":77.6},"dropoff":{"lat":12.95,"lng":77.64},"price":220}`
		req := httptest.NewRequest(http.MethodPost, "/bookings", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("want 201, got %d, body=%s", rr.Code, rr.Body.String())
		}
		var got models.Booking
		if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
			t.Fatalf("json: %v", err)
		}
		if got.BookingID != "b-1" || got.Price != 220 || got.RideStatus != models.RideStatusRequested {
			t.Fatalf("unexpected: %+v", got)
		}
	})

	t.Run("400 validation", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/bookings", strings.NewReader(`{"pickuploc":{"lat":999,"lng":0},"dropoff":{"lat":0,"lng":0},"price":0}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", rr.Code)
		}
	})

	t.Run("400 invalid json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/bookings", strings.NewReader(`{`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", rr.Code)
		}
	})
}

func TestListBookings_Handler(t *testing.T) {
	items := []models.Booking{{BookingID: "b-2", CreatedAt: time.Now().UTC()}}
	h := NewBookingHandler(&fakeBookingService{
		createFn: func(ctx context.Context, in service.CreateBookingInput) (models.Booking, error) {
			return models.Booking{}, nil
		},
		listFn: func(ctx context.Context) ([]models.Booking, error) { return items, nil },
	})
	r := chi.NewRouter()
	h.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/bookings", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rr.Code)
	}
	var got []models.Booking
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("json: %v", err)
	}
	if len(got) != 1 || got[0].BookingID != "b-2" {
		t.Fatalf("unexpected: %+v", got)
	}
}
