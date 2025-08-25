package postgres

import (
	"context"

	"driver_svc/internal/models"
	"driver_svc/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type JobRepoPG struct {
	pool *pgxpool.Pool
}

func NewJobRepo(pool *pgxpool.Pool) *JobRepoPG {
	return &JobRepoPG{pool: pool}
}

// UpsertOpenJob inserts an Open job if it does not already exist (idempotent).
func (r *JobRepoPG) UpsertOpenJob(ctx context.Context, p repository.UpsertJobParams) error {
	const q = `
INSERT INTO jobs
  (booking_id, pickuploc_lat, pickuploc_lng, dropoff_lat, dropoff_lng, price, status)
VALUES
  ($1,$2,$3,$4,$5,$6,'Open')
ON CONFLICT (booking_id) DO NOTHING;
`
	_, err := r.pool.Exec(ctx, q,
		p.BookingID,
		p.PickupLoc.Lat, p.PickupLoc.Lng,
		p.Dropoff.Lat, p.Dropoff.Lng,
		p.Price,
	)
	return err
}

func (r *JobRepoPG) ListOpenJobs(ctx context.Context) ([]models.Job, error) {
	const q = `
SELECT booking_id, pickuploc_lat, pickuploc_lng, dropoff_lat, dropoff_lng, price, status, accepted_driver_id, created_at
FROM jobs
WHERE status = 'Open'
ORDER BY created_at DESC;
`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := make([]models.Job, 0, 32)
	for rows.Next() {
		var j models.Job
		var status string
		if err := rows.Scan(
			&j.BookingID,
			&j.PickupLoc.Lat, &j.PickupLoc.Lng,
			&j.Dropoff.Lat, &j.Dropoff.Lng,
			&j.Price, &status, &j.AcceptedDriverID, &j.CreatedAt,
		); err != nil {
			return nil, err
		}
		j.Status = models.JobStatus(status)
		jobs = append(jobs, j)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return jobs, nil
}

// TryAccept atomically marks a job as Taken if it is currently Open.
// Returns true if this call won (rows affected = 1), false if already taken.
func (r *JobRepoPG) TryAccept(ctx context.Context, bookingID, driverID string) (bool, error) {
	const q = `
UPDATE jobs
SET status = 'Taken', accepted_driver_id = $1
WHERE booking_id = $2 AND status = 'Open';
`
	cmd, err := r.pool.Exec(ctx, q, driverID, bookingID)
	if err != nil {
		return false, err
	}
	return cmd.RowsAffected() == 1, nil
}
