package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Bootstrap(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS drivers (
  driver_id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  is_available BOOLEAN NOT NULL
);`)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS jobs (
  booking_id TEXT PRIMARY KEY,
  pickuploc_lat DOUBLE PRECISION NOT NULL,
  pickuploc_lng DOUBLE PRECISION NOT NULL,
  dropoff_lat DOUBLE PRECISION NOT NULL,
  dropoff_lng DOUBLE PRECISION NOT NULL,
  price INTEGER NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('Open','Taken')) DEFAULT 'Open',
  accepted_driver_id TEXT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);`)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs (status);`)
	return err
}
