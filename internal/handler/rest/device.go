package rest

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/ytobing/devices-api/internal/handler"
	"github.com/ytobing/devices-api/internal/models"
	"github.com/ytobing/devices-api/internal/service"
)

type deviceHandler struct {
	svc    service.DeviceService
	logger *slog.Logger
}

func (h *deviceHandler) getDevices(w http.ResponseWriter, r *http.Request) {
	devices, err := h.svc.GetDevices(r.Context(), r.URL.Query().Get("brand"), r.URL.Query().Get("state"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	resp := []handler.DeviceResponse{}
	for _, d := range devices {
		resp = append(resp, toDeviceResponse(d))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *deviceHandler) getDevice(w http.ResponseWriter, r *http.Request) {
	d, err := h.svc.GetDevice(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toDeviceResponse(d))
}

func (h *deviceHandler) createDevice(w http.ResponseWriter, r *http.Request) {
	var req handler.CreateDeviceRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid JSON body: "+err.Error())
		return
	}

	d, err := h.svc.CreateDevice(r.Context(), req.Name, req.Brand, req.State)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	w.Header().Set("Location", "/devices/"+d.ID)
	writeJSON(w, http.StatusCreated, toDeviceResponse(d))
}

func (h *deviceHandler) replaceDevice(w http.ResponseWriter, r *http.Request) {
	var req handler.UpdateDeviceRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid JSON body: "+err.Error())
		return
	}
	if req.Name == nil || req.Brand == nil || req.State == nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "name, brand and state are required")
		return
	}
	h.updateDevice(w, r, req)
}

func (h *deviceHandler) patchDevice(w http.ResponseWriter, r *http.Request) {
	var req handler.UpdateDeviceRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "invalid JSON body: "+err.Error())
		return
	}
	if req.Name == nil && req.Brand == nil && req.State == nil {
		writeError(w, http.StatusBadRequest, "invalid_input", "at least one of name, brand or state is required")
		return
	}
	h.updateDevice(w, r, req)
}

func (h *deviceHandler) updateDevice(w http.ResponseWriter, r *http.Request, req handler.UpdateDeviceRequest) {
	d, err := h.svc.UpdateDevice(r.Context(), chi.URLParam(r, "id"), models.DeviceUpdate{
		Name:  req.Name,
		Brand: req.Brand,
		State: req.State,
	})
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toDeviceResponse(d))
}

func (h *deviceHandler) deleteDevice(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteDevice(r.Context(), chi.URLParam(r, "id")); err != nil {
		h.writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *deviceHandler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, models.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid_input", err.Error())
	case errors.Is(err, models.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", err.Error())
	case errors.Is(err, models.ErrDeviceInUse):
		writeError(w, http.StatusConflict, "conflict", err.Error())
	default:
		h.logger.Error("request failed", "err", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
