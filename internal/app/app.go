package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ytobing/devices-api/internal/handler/rest"
	"github.com/ytobing/devices-api/internal/notifier"
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

	broker := notifier.New()
	repo := repository.NewRepository(db)
	service := service.NewDeviceService(repo, broker)
	restHandler := rest.NewHandler(service, broker, logger)

	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: restHandler,
	}
	srv.RegisterOnShutdown(broker.Close)

	context.AfterFunc(ctx, func() { srv.Shutdown(context.Background()) })

	logger.Info("http server listening", "addr", cfg.HTTPAddr)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
