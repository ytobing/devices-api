package rest

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/ytobing/devices-api/internal/service"
)

func NewHandler(devices service.DeviceService, logger *slog.Logger) http.Handler {
	h := &deviceHandler{svc: devices, logger: logger}

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.Recoverer)

	r.Route("/devices", func(r chi.Router) {
		r.Get("/", h.getDevices)
	})
	return r
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
