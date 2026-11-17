package app

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"

	"github.com/ytobing/devices-api/migrations"
)

var DBPath = "devices.db"

var HTTPAddr = ":8080"

type Config struct {
	DBPath   string
	HTTPAddr string
}

func Load() (Config, error) {
	return Config{
		DBPath:   DBPath,
		HTTPAddr: HTTPAddr,
	}, nil
}

func InitSQLite(ctx context.Context, path string) (*sqlx.DB, error) {
	db, err := sqlx.ConnectContext(ctx, "sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("connect sqlite: %w", err)
	}

	if err := migrate(ctx, db.DB); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func migrate(ctx context.Context, db *sql.DB) error {
	provider, err := goose.NewProvider(goose.DialectSQLite3, db, migrations.FS)
	if err != nil {
		return fmt.Errorf("init migrations: %w", err)
	}
	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil
}
