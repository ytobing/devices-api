package app

import (
	"context"
	"fmt"
	"log/slog"
)

func Run(ctx context.Context, cfg Config, logger *slog.Logger) error {
	db, err := InitSQLite(ctx, cfg.DBPath)
	if err != nil {
		return fmt.Errorf("init db: %w", err)
	}
	defer db.Close()
	logger.Info("database ready", "path", cfg.DBPath)

	return nil
}
