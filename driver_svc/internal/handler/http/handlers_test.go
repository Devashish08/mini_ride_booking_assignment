package handlerhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"driver_svc/internal/models"
	"driver_svc/internal/service"

	"github.com/go-chi/chi/v5"
)

type fakeJobsService struct {
	listDriversFn  func(ctx context.Context) ([]models.Driver, error)
	listOpenJobsFn func(ctx context.Context) ([]models.Job, error)
	acceptFn       func(ctx context.Context, bookingID string, driverID string) error
}

func (f *fakeJobsService) ListDrivers(ctx context.Context) ([]models.Driver, error) {
	return f.listDriversFn(ctx)
}
func (f *fakeJobsService) ListOpenJobs(ctx context.Context) ([]models.Job, error) {
	return f.listOpenJobsFn(ctx)
}
func (f *fakeJobsService) AcceptJob(ctx context.Context, b, d string) error {
	return f.acceptFn(ctx, b, d)
}

func setup(t *testing.T, svc *fakeJobsService) *chi.Mux {
	t.Helper()
	h := NewJobsHandler(svc)
	r := chi.NewRouter()
	h.RegisterRoutes(r)
	return r
}

func TestListDrivers(t *testing.T) {
	r := setup(t, &fakeJobsService{
		listDriversFn: func(ctx context.Context) ([]models.Driver, error) {
			return []models.Driver{{DriverID: "d-1", Name: "Asha", IsAvailable: true}}, nil
		},
		listOpenJobsFn: func(ctx context.Context) ([]models.Job, error) { return nil, nil },
		acceptFn:       func(ctx context.Context, b, d string) error { return nil },
	})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/drivers", nil)
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rr.Code)
	}
	var got []models.Driver
	_ = json.Unmarshal(rr.Body.Bytes(), &got)
	if len(got) != 1 || got[0].DriverID != "d-1" {
		t.Fatalf("unexpected: %+v", got)
	}
}

func TestListJobs(t *testing.T) {
	r := setup(t, &fakeJobsService{
		listDriversFn: func(ctx context.Context) ([]models.Driver, error) { return nil, nil },
		listOpenJobsFn: func(ctx context.Context) ([]models.Job, error) {
			return []models.Job{{BookingID: "b-1"}}, nil
		},
		acceptFn: func(ctx context.Context, b, d string) error { return nil },
	})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rr.Code)
	}
}

func TestAcceptJob_Table(t *testing.T) {
	cases := []struct {
		name       string
		body       string
		err        error
		wantStatus int
	}{
		{"ok", `{"driver_id":"d-1"}`, nil, http.StatusOK},
		{"missing driver_id", `{}`, nil, http.StatusBadRequest},
		{"invalid json", `{`, nil, http.StatusBadRequest},
		{"driver not found", `{"driver_id":"x"}`, service.ErrDriverNotFound, http.StatusNotFound},
		{"already taken", `{"driver_id":"d-2"}`, service.ErrJobAlreadyTaken, http.StatusConflict},
		{"generic", `{"driver_id":"d-1"}`, context.Canceled, http.StatusInternalServerError},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := setup(t, &fakeJobsService{
				listDriversFn:  func(ctx context.Context) ([]models.Driver, error) { return nil, nil },
				listOpenJobsFn: func(ctx context.Context) ([]models.Job, error) { return nil, nil },
				acceptFn:       func(ctx context.Context, b, d string) error { return c.err },
			})
			var body *strings.Reader
			if c.name == "invalid json" {
				body = strings.NewReader(c.body)
			} else {
				body = strings.NewReader(c.body)
			}
			req := httptest.NewRequest(http.MethodPost, "/jobs/b-1/accept", body)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)
			if rr.Code != c.wantStatus {
				t.Fatalf("want %d, got %d, body=%s", c.wantStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestAcceptJob_UnknownField(t *testing.T) {
	r := setup(t, &fakeJobsService{
		listDriversFn:  func(ctx context.Context) ([]models.Driver, error) { return nil, nil },
		listOpenJobsFn: func(ctx context.Context) ([]models.Job, error) { return nil, nil },
		acceptFn:       func(ctx context.Context, b, d string) error { return nil },
	})
	req := httptest.NewRequest(http.MethodPost, "/jobs/b-1/accept", bytes.NewBufferString(`{"driver_id":"d-1","extra":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rr.Code)
	}
}
