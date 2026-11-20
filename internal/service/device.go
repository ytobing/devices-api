package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/ytobing/devices-api/internal/models"
	"github.com/ytobing/devices-api/internal/repository"
)

type DeviceService interface {
	GetDevices(ctx context.Context, brand, state string) ([]models.Device, error)
	GetDevice(ctx context.Context, id string) (models.Device, error)
	CreateDevice(ctx context.Context, name, brand, state string) (models.Device, error)
	UpdateDevice(ctx context.Context, id string, upd models.DeviceUpdate) (models.Device, error)
	DeleteDevice(ctx context.Context, id string) error
}

type Notifier interface {
	Publish(change models.StateChange)
}

type deviceService struct {
	repo     repository.Repository
	notifier Notifier
}

func NewDeviceService(repo repository.Repository, notifier Notifier) DeviceService {
	return &deviceService{
		repo:     repo,
		notifier: notifier,
	}
}

var errInvalidState = fmt.Errorf("%w: state must be one of: AVAILABLE, IN_USE, INACTIVE", models.ErrInvalidInput)

func (s *deviceService) GetDevices(ctx context.Context, brand, state string) ([]models.Device, error) {
	if state != "" && !models.IsValidState(state) {
		return nil, errInvalidState
	}
	return s.repo.GetDevices(ctx, brand, state)
}

func (s *deviceService) GetDevice(ctx context.Context, id string) (models.Device, error) {
	return s.repo.GetDeviceByID(ctx, id)
}

func (s *deviceService) CreateDevice(ctx context.Context, name, brand, state string) (models.Device, error) {
	if state == "" {
		state = models.StateAvailable
	}
	if err := validate(name, brand, state); err != nil {
		return models.Device{}, err
	}
	return s.repo.CreateDevice(ctx, models.Device{
		ID:    uuid.NewString(),
		Name:  name,
		Brand: brand,
		State: state,
	})
}

func (s *deviceService) UpdateDevice(ctx context.Context, id string, upd models.DeviceUpdate) (models.Device, error) {
	current, err := s.repo.GetDeviceByID(ctx, id)
	if err != nil {
		return models.Device{}, err
	}

	updated := current
	if upd.Name != nil {
		updated.Name = *upd.Name
	}
	if upd.Brand != nil {
		updated.Brand = *upd.Brand
	}
	if upd.State != nil {
		updated.State = *upd.State
	}

	if err := validate(updated.Name, updated.Brand, updated.State); err != nil {
		return models.Device{}, err
	}
	if current.State == models.StateInUse && (updated.Name != current.Name || updated.Brand != current.Brand) {
		return models.Device{}, fmt.Errorf("%w: name and brand cannot be updated", models.ErrDeviceInUse)
	}

	saved, err := s.repo.UpdateDevice(ctx, updated)
	if err != nil {
		return models.Device{}, err
	}

	if saved.State != current.State {
		s.notifier.Publish(models.StateChange{
			DeviceID:      saved.ID,
			PreviousState: current.State,
			NewState:      saved.State,
		})
	}
	return saved, nil
}

func (s *deviceService) DeleteDevice(ctx context.Context, id string) error {
	current, err := s.repo.GetDeviceByID(ctx, id)
	if err != nil {
		return err
	}
	if current.State == models.StateInUse {
		return fmt.Errorf("%w: it cannot be deleted", models.ErrDeviceInUse)
	}
	return s.repo.DeleteDevice(ctx, id)
}

func validate(name, brand, state string) error {
	if strings.TrimSpace(name) == "" || strings.TrimSpace(brand) == "" {
		return fmt.Errorf("%w: name and brand are required", models.ErrInvalidInput)
	}
	if !models.IsValidState(state) {
		return errInvalidState
	}
	return nil
}
