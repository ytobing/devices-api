package rest

import (
	"log/slog"
	"net/http"

	"github.com/ytobing/devices-api/internal/service"
)

type deviceHandler struct {
	svc    service.DeviceService
	logger *slog.Logger
}

var validStates = map[string]bool{"AVAILABLE": true, "IN_USE": true, "INACTIVE": true}

type deviceResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Brand     string `json:"brand"`
	State     string `json:"state"`
	CreatedAt string `json:"created_at"`
}

func (h *deviceHandler) getDevices(w http.ResponseWriter, r *http.Request) {
	brand := r.URL.Query().Get("brand")
	state := r.URL.Query().Get("state")
	if state != "" && !validStates[state] {
		writeJSON(w, http.StatusBadRequest, errorResponse{
			Code:    "invalid_input",
			Message: "state must be one of: AVAILABLE, IN_USE, INACTIVE",
		})
		return
	}

	devices, err := h.svc.GetDevices(r.Context(), brand, state)
	if err != nil {
		h.logger.Error("get devices", "err", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Code:    "internal_error",
			Message: "internal server error",
		})
		return
	}

	resp := make([]deviceResponse, 0, len(devices))
	for _, d := range devices {
		resp = append(resp, deviceResponse{
			ID:        d.ID,
			Name:      d.Name,
			Brand:     d.Brand,
			State:     d.State,
			CreatedAt: d.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, resp)
}
