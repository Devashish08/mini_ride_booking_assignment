package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Bootstrap(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS bookings (
  booking_id TEXT PRIMARY KEY,
  pickuploc_lat DOUBLE PRECISION NOT NULL,
  pickuploc_lng DOUBLE PRECISION NOT NULL,
  dropoff_lat DOUBLE PRECISION NOT NULL,
  dropoff_lng DOUBLE PRECISION NOT NULL,
  price INTEGER NOT NULL,
  ride_status TEXT NOT NULL CHECK (ride_status IN ('Requested','Accepted')),
  driver_id TEXT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);`)
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_bookings_created_at ON bookings (created_at DESC);`)
	return err
}
