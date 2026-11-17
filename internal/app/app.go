package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ytobing/devices-api/internal/handler/rest"
	"github.com/ytobing/devices-api/internal/repository"
	"github.com/ytobing/devices-api/internal/service"
)

func Run(ctx context.Context, cfg Config, logger *slog.Logger) error {
	db, err := InitSQLite(ctx, cfg.DBPath)
	if err != nil {
		return fmt.Errorf("init db: %w", err)
	}
	defer db.Close()
	logger.Info("database ready", "path", cfg.DBPath)

	repo := repository.NewRepository(db)
	service := service.NewDeviceService(repo)
	restHandler := rest.NewHandler(service, logger)

	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: restHandler,
	}

	context.AfterFunc(ctx, func() { srv.Shutdown(context.Background()) })

	logger.Info("http server listening", "addr", cfg.HTTPAddr)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
