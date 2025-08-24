package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Params struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func Connect(ctx context.Context, p Params) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		p.User, p.Password, p.Host, p.Port, p.Name)

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 5

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return pgxpool.NewWithConfig(ctx, cfg)
}
