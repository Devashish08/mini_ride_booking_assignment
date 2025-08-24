package seed

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func SeedDrivers(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
INSERT INTO drivers (driver_id, name, is_available)
VALUES ('d-1','Asha', TRUE)
ON CONFLICT (driver_id) DO UPDATE
SET name = EXCLUDED.name, is_available = EXCLUDED.is_available;`)
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx, `
INSERT INTO drivers (driver_id, name, is_available)
VALUES ('d-2','Ravi', TRUE)
ON CONFLICT (driver_id) DO UPDATE
SET name = EXCLUDED.name, is_available = EXCLUDED.is_available;`)
	return err
}
