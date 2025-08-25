package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"driver_svc/internal/events"
	"driver_svc/internal/models"
	"driver_svc/internal/repository"
)

// fakes

type fakeDriverRepo struct {
	getFn  func(ctx context.Context, driverID string) (models.Driver, bool, error)
	listFn func(ctx context.Context) ([]models.Driver, error)
}

func (f *fakeDriverRepo) ListAll(ctx context.Context) ([]models.Driver, error) {
	if f.listFn != nil {
		return f.listFn(ctx)
	}
	return nil, nil
}
func (f *fakeDriverRepo) GetByID(ctx context.Context, driverID string) (models.Driver, bool, error) {
	if f.getFn != nil {
		return f.getFn(ctx, driverID)
	}
	return models.Driver{}, false, nil
}

type fakeJobRepo struct {
	tryFn    func(ctx context.Context, bookingID, driverID string) (bool, error)
	upsertFn func(ctx context.Context, p repository.UpsertJobParams) error
	listFn   func(ctx context.Context) ([]models.Job, error)
}

func (f *fakeJobRepo) UpsertOpenJob(ctx context.Context, p repository.UpsertJobParams) error {
	if f.upsertFn != nil {
		return f.upsertFn(ctx, p)
	}
	return nil
}
func (f *fakeJobRepo) ListOpenJobs(ctx context.Context) ([]models.Job, error) {
	if f.listFn != nil {
		return f.listFn(ctx)
	}
	return nil, nil
}
func (f *fakeJobRepo) TryAccept(ctx context.Context, bookingID, driverID string) (bool, error) {
	if f.tryFn != nil {
		return f.tryFn(ctx, bookingID, driverID)
	}
	return false, nil
}

type fakeProducer struct {
	mu     sync.Mutex
	calls  int
	events []events.BookingAccepted
	err    error
}

func (p *fakeProducer) ProduceBookingAccepted(ctx context.Context, evt events.BookingAccepted) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls++
	p.events = append(p.events, evt)
	return p.err
}

// table-driven tests

func TestAcceptJob_Table(t *testing.T) {
	tests := []struct {
		name      string
		driverOK  bool
		available bool
		tryAccept func(ctx context.Context, bID, dID string) (bool, error)
		prodErr   error
		wantErr   error
		wantCalls int
	}{
		{
			name:     "success -> event produced",
			driverOK: true, available: true,
			tryAccept: func(_ context.Context, _, _ string) (bool, error) { return true, nil },
			wantErr:   nil, wantCalls: 1,
		},
		{
			name:     "already taken -> 409",
			driverOK: true, available: true,
			tryAccept: func(_ context.Context, _, _ string) (bool, error) { return false, nil },
			wantErr:   ErrJobAlreadyTaken, wantCalls: 0,
		},
		{
			name:     "driver missing -> 404",
			driverOK: false, available: false,
			tryAccept: func(_ context.Context, _, _ string) (bool, error) { return false, nil },
			wantErr:   ErrDriverNotFound, wantCalls: 0,
		},
		{
			name:     "repo error -> 500",
			driverOK: true, available: true,
			tryAccept: func(_ context.Context, _, _ string) (bool, error) { return false, errors.New("db err") },
			wantErr:   errors.New("db err"), wantCalls: 0,
		},
		{
			name:     "producer error -> 500",
			driverOK: true, available: true,
			tryAccept: func(_ context.Context, _, _ string) (bool, error) { return true, nil },
			prodErr:   errors.New("produce err"),
			wantErr:   errors.New("produce err"), wantCalls: 1, // tried to produce once
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dr := &fakeDriverRepo{
				getFn: func(ctx context.Context, driverID string) (models.Driver, bool, error) {
					return models.Driver{DriverID: driverID, IsAvailable: tc.available}, tc.driverOK, nil
				},
			}
			jr := &fakeJobRepo{tryFn: tc.tryAccept}
			prod := &fakeProducer{err: tc.prodErr}

			svc := NewJobsService(dr, jr, prod, nil)
			err := svc.AcceptJob(context.Background(), "b-1", "d-1")

			if (tc.wantErr == nil) != (err == nil) {
				t.Fatalf("wantErr=%v got=%v", tc.wantErr, err)
			}
			if tc.wantErr != nil && err != nil {
				// compare error types where applicable
				if !errors.Is(err, tc.wantErr) && err.Error() != tc.wantErr.Error() {
					t.Fatalf("want %v got %v", tc.wantErr, err)
				}
			}
			if prod.calls != tc.wantCalls {
				t.Fatalf("producer calls: want %d got %d", tc.wantCalls, prod.calls)
			}
		})
	}
}

func TestAcceptJob_FirstWriterWins_Concurrent(t *testing.T) {
	var once sync.Once
	jr := &fakeJobRepo{
		tryFn: func(_ context.Context, _, _ string) (bool, error) {
			won := false
			once.Do(func() { won = true })
			return won, nil
		},
	}
	dr := &fakeDriverRepo{
		getFn: func(ctx context.Context, driverID string) (models.Driver, bool, error) {
			return models.Driver{DriverID: driverID, IsAvailable: true}, true, nil
		},
	}
	prod := &fakeProducer{}
	svc := NewJobsService(dr, jr, prod, nil)

	var wg sync.WaitGroup
	errs := make([]error, 2)
	wg.Add(2)
	go func() { defer wg.Done(); errs[0] = svc.AcceptJob(context.Background(), "b-1", "d-1") }()
	go func() { defer wg.Done(); errs[1] = svc.AcceptJob(context.Background(), "b-1", "d-2") }()
	wg.Wait()

	var success, conflict int
	for _, e := range errs {
		if e == nil {
			success++
		} else if errors.Is(e, ErrJobAlreadyTaken) {
			conflict++
		} else {
			t.Fatalf("unexpected error: %v", e)
		}
	}
	if success != 1 || conflict != 1 {
		t.Fatalf("want 1 success, 1 conflict; got success=%d conflict=%d", success, conflict)
	}
	// exactly one event produced
	time.Sleep(10 * time.Millisecond)
	if prod.calls != 1 {
		t.Fatalf("producer calls want 1 got %d", prod.calls)
	}
}
