package repository

import (
	"context"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	GetDevices(ctx context.Context, brand, state string) ([]Device, error)
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

type repository struct {
	db *sqlx.DB
}

type Device struct {
	ID        string `db:"id"`
	Name      string `db:"name"`
	Brand     string `db:"brand"`
	State     string `db:"state"`
	CreatedAt string `db:"created_at"`
}

func (r *repository) GetDevices(ctx context.Context, brand, state string) ([]Device, error) {
	query := GetDevicesQuery
	var conds []string
	var args []any

	if brand != "" {
		conds = append(conds, "brand = ?")
		args = append(args, brand)
	}
	if state != "" {
		conds = append(conds, "state = ?")
		args = append(args, state)
	}
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}

	var devices []Device
	if err := r.db.SelectContext(ctx, &devices, query, args...); err != nil {
		return nil, err
	}
	return devices, nil
}
