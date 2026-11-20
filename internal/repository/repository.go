package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/ytobing/devices-api/internal/models"
)

type Repository interface {
	GetDevices(ctx context.Context, brand, state string) ([]models.Device, error)
	GetDeviceByID(ctx context.Context, id string) (models.Device, error)
	CreateDevice(ctx context.Context, d models.Device) (models.Device, error)
	UpdateDevice(ctx context.Context, d models.Device) (models.Device, error)
	DeleteDevice(ctx context.Context, id string) error
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

type repository struct {
	db *sqlx.DB
}

func (r *repository) GetDevices(ctx context.Context, brand, state string) ([]models.Device, error) {
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

	var devices []models.Device
	if err := r.db.SelectContext(ctx, &devices, query, args...); err != nil {
		return nil, err
	}
	return devices, nil
}

func (r *repository) GetDeviceByID(ctx context.Context, id string) (models.Device, error) {
	var d models.Device
	err := r.db.GetContext(ctx, &d, GetDeviceByIDQuery, id)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Device{}, models.ErrNotFound
	}
	return d, err
}

func (r *repository) CreateDevice(ctx context.Context, d models.Device) (models.Device, error) {
	var created models.Device
	err := r.db.GetContext(ctx, &created, CreateDeviceQuery, d.ID, d.Name, d.Brand, d.State)
	return created, err
}

func (r *repository) UpdateDevice(ctx context.Context, d models.Device) (models.Device, error) {
	var updated models.Device
	err := r.db.GetContext(ctx, &updated, UpdateDeviceQuery, d.Name, d.Brand, d.State, d.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Device{}, models.ErrNotFound
	}
	return updated, err
}

func (r *repository) DeleteDevice(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, DeleteDeviceQuery, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return models.ErrNotFound
	}
	return nil
}
