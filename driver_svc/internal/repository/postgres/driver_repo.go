package postgres

import (
	"context"

	"driver_svc/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DriverRepoPG struct {
	pool *pgxpool.Pool
}

func NewDriverRepo(pool *pgxpool.Pool) *DriverRepoPG {
	return &DriverRepoPG{pool: pool}
}

func (r *DriverRepoPG) ListAll(ctx context.Context) ([]models.Driver, error) {
	const q = `
SELECT driver_id, name, is_available
FROM drivers
ORDER BY driver_id;
`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	drivers := make([]models.Driver, 0, 16)
	for rows.Next() {
		var d models.Driver
		if err := rows.Scan(&d.DriverID, &d.Name, &d.IsAvailable); err != nil {
			return nil, err
		}
		drivers = append(drivers, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return drivers, nil
}

func (r *DriverRepoPG) GetByID(ctx context.Context, driverID string) (models.Driver, bool, error) {
	const q = `SELECT driver_id, name, is_available FROM drivers WHERE driver_id = $1;`
	var d models.Driver
	if err := r.pool.QueryRow(ctx, q, driverID).Scan(&d.DriverID, &d.Name, &d.IsAvailable); err != nil {
		if err == pgx.ErrNoRows {
			return models.Driver{}, false, nil
		}
		return models.Driver{}, false, err
	}
	return d, true, nil
}
