package repository

import "github.com/jmoiron/sqlx"

type Repository interface {
	GetDevices() ([]Device, error)
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{}
}

type repository struct {
	db *sqlx.DB
}

type Device struct {
	ID        string
	Name      string
	Brand     string
	State     string
	CreatedAt string
}

func (r *repository) GetDevices() ([]Device, error) {
	return nil, nil
}
