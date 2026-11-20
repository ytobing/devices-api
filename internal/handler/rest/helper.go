package rest

import (
	"github.com/ytobing/devices-api/internal/handler"
	"github.com/ytobing/devices-api/internal/models"
)

func toDeviceResponse(d models.Device) handler.DeviceResponse {
	return handler.DeviceResponse{
		ID:        d.ID,
		Name:      d.Name,
		Brand:     d.Brand,
		State:     d.State,
		CreatedAt: d.CreatedAt,
	}
}
