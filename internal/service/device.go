package service

import "github.com/ytobing/devices-api/internal/repository"

type DeviceService interface {
	GetDevices() ([]Device, error)
}

type deviceService struct {
	repo repository.Repository
}

func NewDeviceService(repo repository.Repository) DeviceService {
	return &deviceService{
		repo: repo,
	}
}

type Device struct {
	ID        string
	Name      string
	Brand     string
	State     string
	CreatedAt string
}

func (s *deviceService) GetDevices() ([]Device, error) {
	rows, err := s.repo.GetDevices()
	if err != nil {
		return nil, err
	}

	devices := make([]Device, 0, len(rows))
	for _, d := range rows {
		devices = append(devices, Device{
			ID:        d.ID,
			Name:      d.Name,
			Brand:     d.Brand,
			State:     d.State,
			CreatedAt: d.CreatedAt,
		})
	}
	return devices, nil
}
