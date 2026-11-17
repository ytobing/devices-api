package app

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"

	"github.com/ytobing/devices-api/migrations"
)

var DBPath = "devices.db"

type Config struct {
	DBPath string
}

func Load() (Config, error) {
	return Config{
		DBPath: DBPath,
	}, nil
}

func InitSQLite(ctx context.Context, path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	db.SetMaxOpenConns(1)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	if err := migrate(ctx, db); err != nil {
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
