package rest

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/ytobing/devices-api/internal/service"
)

func NewHandler(devices service.DeviceService, events http.Handler, logger *slog.Logger) http.Handler {
	h := &deviceHandler{svc: devices, logger: logger}

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.Recoverer)

	r.Get("/docs", serveDocs)
	r.Get("/docs/openapi.yaml", serveSpec)

	r.Route("/devices", func(r chi.Router) {
		r.Get("/", h.getDevices)
		r.Post("/", h.createDevice)
		r.Method(http.MethodGet, "/events", events)
		r.Get("/{id}", h.getDevice)
		r.Put("/{id}", h.replaceDevice)
		r.Patch("/{id}", h.patchDevice)
		r.Delete("/{id}", h.deleteDevice)
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

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorResponse{Code: code, Message: message})
}

func decodeJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
