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

type deviceResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Brand     string `json:"brand"`
	State     string `json:"state"`
	CreatedAt string `json:"created_at"`
}

func (h *deviceHandler) getDevices(w http.ResponseWriter, r *http.Request) {
	devices, err := h.svc.GetDevices()
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
