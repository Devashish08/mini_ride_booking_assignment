package postgres

import (
	"context"

	"booking_svc/internal/models"
	"booking_svc/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type BookingRepoPG struct {
	pool *pgxpool.Pool
}

func NewBookingRepo(pool *pgxpool.Pool) *BookingRepoPG {
	return &BookingRepoPG{pool: pool}
}

func (r *BookingRepoPG) Create(ctx context.Context, p repository.CreateBookingParams) (models.Booking, error) {
	const q = `
INSERT INTO bookings
  (booking_id, pickuploc_lat, pickuploc_lng, dropoff_lat, dropoff_lng, price, ride_status, driver_id)
VALUES
  ($1,$2,$3,$4,$5,$6,$7,$8)
RETURNING booking_id, pickuploc_lat, pickuploc_lng, dropoff_lat, dropoff_lng, price, ride_status, driver_id, created_at;
`
	row := r.pool.QueryRow(ctx, q,
		p.BookingID,
		p.PickupLoc.Lat, p.PickupLoc.Lng,
		p.Dropoff.Lat, p.Dropoff.Lng,
		p.Price, string(p.RideStatus), p.DriverID,
	)

	var b models.Booking
	var status string
	if err := row.Scan(
		&b.BookingID,
		&b.PickupLoc.Lat, &b.PickupLoc.Lng,
		&b.Dropoff.Lat, &b.Dropoff.Lng,
		&b.Price, &status, &b.DriverID, &b.CreatedAt,
	); err != nil {
		return models.Booking{}, err
	}
	b.RideStatus = models.RideStatus(status)
	return b, nil
}

func (r *BookingRepoPG) ListAll(ctx context.Context) ([]models.Booking, error) {
	const q = `
SELECT booking_id, pickuploc_lat, pickuploc_lng, dropoff_lat, dropoff_lng, price, ride_status, driver_id, created_at
FROM bookings
ORDER BY created_at DESC;
`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bookings := make([]models.Booking, 0, 32)
	for rows.Next() {
		var b models.Booking
		var status string
		if err := rows.Scan(
			&b.BookingID,
			&b.PickupLoc.Lat, &b.PickupLoc.Lng,
			&b.Dropoff.Lat, &b.Dropoff.Lng,
			&b.Price, &status, &b.DriverID, &b.CreatedAt,
		); err != nil {
			return nil, err
		}
		b.RideStatus = models.RideStatus(status)
		bookings = append(bookings, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return bookings, nil
}

func (r *BookingRepoPG) MarkAccepted(ctx context.Context, bookingID, driverID string) (bool, error) {
	const q = `
UPDATE bookings
SET ride_status = 'Accepted', driver_id = $1
WHERE booking_id = $2 AND ride_status = 'Requested';
`
	cmd, err := r.pool.Exec(ctx, q, driverID, bookingID)
	if err != nil {
		return false, err
	}
	return cmd.RowsAffected() == 1, nil
}
