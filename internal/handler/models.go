package handler

type DeviceResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Brand     string `json:"brand"`
	State     string `json:"state"`
	CreatedAt string `json:"created_at"`
}

type CreateDeviceRequest struct {
	Name  string `json:"name"`
	Brand string `json:"brand"`
	State string `json:"state"`
}

type UpdateDeviceRequest struct {
	Name  *string `json:"name"`
	Brand *string `json:"brand"`
	State *string `json:"state"`
}
