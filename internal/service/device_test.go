package service

import (
	"context"
	"errors"
	"testing"

	"github.com/ytobing/devices-api/internal/models"
)

type fakeRepo struct {
	devices   map[string]models.Device
	updateErr error
}

func newFakeRepo(devices ...models.Device) *fakeRepo {
	r := &fakeRepo{devices: map[string]models.Device{}}
	for _, d := range devices {
		r.devices[d.ID] = d
	}
	return r
}

func (r *fakeRepo) GetDevices(_ context.Context, brand, state string) ([]models.Device, error) {
	var out []models.Device
	for _, d := range r.devices {
		if (brand == "" || d.Brand == brand) && (state == "" || d.State == state) {
			out = append(out, d)
		}
	}
	return out, nil
}

func (r *fakeRepo) GetDeviceByID(_ context.Context, id string) (models.Device, error) {
	d, ok := r.devices[id]
	if !ok {
		return models.Device{}, models.ErrNotFound
	}
	return d, nil
}

func (r *fakeRepo) CreateDevice(_ context.Context, d models.Device) (models.Device, error) {
	d.CreatedAt = "2026-11-20T10:00:00Z"
	r.devices[d.ID] = d
	return d, nil
}

func (r *fakeRepo) UpdateDevice(_ context.Context, d models.Device) (models.Device, error) {
	if r.updateErr != nil {
		return models.Device{}, r.updateErr
	}
	if _, ok := r.devices[d.ID]; !ok {
		return models.Device{}, models.ErrNotFound
	}
	r.devices[d.ID] = d
	return d, nil
}

func (r *fakeRepo) DeleteDevice(_ context.Context, id string) error {
	if _, ok := r.devices[id]; !ok {
		return models.ErrNotFound
	}
	delete(r.devices, id)
	return nil
}

type fakeNotifier struct {
	changes []models.StateChange
}

func (n *fakeNotifier) Publish(c models.StateChange) {
	n.changes = append(n.changes, c)
}

func ptr(s string) *string { return &s }

func device(state string) models.Device {
	return models.Device{ID: "d1", Name: "iPhone", Brand: "Apple", State: state, CreatedAt: "2026-11-20T09:00:00Z"}
}

func TestCreateDeviceDefaultsToAvailable(t *testing.T) {
	svc := NewDeviceService(newFakeRepo(), &fakeNotifier{})

	d, err := svc.CreateDevice(context.Background(), "iPhone", "Apple", "")
	if err != nil {
		t.Fatal(err)
	}
	if d.State != models.StateAvailable || d.ID == "" {
		t.Fatalf("unexpected device: %+v", d)
	}
}

func TestCreateDeviceValidation(t *testing.T) {
	tests := map[string]struct{ name, brand, state string }{
		"missing name":  {"", "Apple", ""},
		"missing brand": {"iPhone", " ", ""},
		"invalid state": {"iPhone", "Apple", "broken"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			svc := NewDeviceService(newFakeRepo(), &fakeNotifier{})
			_, err := svc.CreateDevice(context.Background(), tc.name, tc.brand, tc.state)
			if !errors.Is(err, models.ErrInvalidInput) {
				t.Fatalf("want ErrInvalidInput, got %v", err)
			}
		})
	}
}

func TestGetDevicesInvalidState(t *testing.T) {
	svc := NewDeviceService(newFakeRepo(), &fakeNotifier{})
	if _, err := svc.GetDevices(context.Background(), "", "broken"); !errors.Is(err, models.ErrInvalidInput) {
		t.Fatalf("want ErrInvalidInput, got %v", err)
	}
}

func TestUpdateDevicePartialKeepsOtherFieldsAndNotifies(t *testing.T) {
	n := &fakeNotifier{}
	svc := NewDeviceService(newFakeRepo(device(models.StateAvailable)), n)

	d, err := svc.UpdateDevice(context.Background(), "d1", models.DeviceUpdate{State: ptr(models.StateInUse)})
	if err != nil {
		t.Fatal(err)
	}
	if d.Name != "iPhone" || d.Brand != "Apple" || d.State != models.StateInUse || d.CreatedAt != "2026-11-20T09:00:00Z" {
		t.Fatalf("unexpected device: %+v", d)
	}
	want := models.StateChange{DeviceID: "d1", PreviousState: models.StateAvailable, NewState: models.StateInUse}
	if len(n.changes) != 1 || n.changes[0] != want {
		t.Fatalf("want one notification %+v, got %+v", want, n.changes)
	}
}

func TestUpdateDeviceWithoutStateChangeDoesNotNotify(t *testing.T) {
	n := &fakeNotifier{}
	svc := NewDeviceService(newFakeRepo(device(models.StateAvailable)), n)

	if _, err := svc.UpdateDevice(context.Background(), "d1", models.DeviceUpdate{Name: ptr("iPhone 16")}); err != nil {
		t.Fatal(err)
	}
	if len(n.changes) != 0 {
		t.Fatalf("want no notification, got %+v", n.changes)
	}
}

func TestUpdateDeviceInUseCannotChangeNameOrBrand(t *testing.T) {
	tests := map[string]models.DeviceUpdate{
		"name":  {Name: ptr("iPhone 16")},
		"brand": {Brand: ptr("Samsung")},
	}
	for name, upd := range tests {
		t.Run(name, func(t *testing.T) {
			repo := newFakeRepo(device(models.StateInUse))
			n := &fakeNotifier{}
			svc := NewDeviceService(repo, n)

			_, err := svc.UpdateDevice(context.Background(), "d1", upd)
			if !errors.Is(err, models.ErrDeviceInUse) {
				t.Fatalf("want ErrDeviceInUse, got %v", err)
			}
			if repo.devices["d1"] != device(models.StateInUse) {
				t.Fatalf("device was modified: %+v", repo.devices["d1"])
			}
			if len(n.changes) != 0 {
				t.Fatalf("want no notification, got %+v", n.changes)
			}
		})
	}
}

func TestUpdateDeviceInUseCanChangeStateWithSameNameAndBrand(t *testing.T) {
	n := &fakeNotifier{}
	svc := NewDeviceService(newFakeRepo(device(models.StateInUse)), n)

	d, err := svc.UpdateDevice(context.Background(), "d1", models.DeviceUpdate{
		Name:  ptr("iPhone"),
		Brand: ptr("Apple"),
		State: ptr(models.StateAvailable),
	})
	if err != nil {
		t.Fatal(err)
	}
	if d.State != models.StateAvailable || len(n.changes) != 1 {
		t.Fatalf("unexpected result: %+v, notifications: %+v", d, n.changes)
	}
}

func TestUpdateDeviceDoesNotNotifyWhenPersistFails(t *testing.T) {
	repo := newFakeRepo(device(models.StateAvailable))
	repo.updateErr = errors.New("db down")
	n := &fakeNotifier{}
	svc := NewDeviceService(repo, n)

	if _, err := svc.UpdateDevice(context.Background(), "d1", models.DeviceUpdate{State: ptr(models.StateInUse)}); err == nil {
		t.Fatal("want error")
	}
	if len(n.changes) != 0 {
		t.Fatalf("want no notification, got %+v", n.changes)
	}
}

func TestUpdateDeviceValidation(t *testing.T) {
	svc := NewDeviceService(newFakeRepo(device(models.StateAvailable)), &fakeNotifier{})

	if _, err := svc.UpdateDevice(context.Background(), "d1", models.DeviceUpdate{Name: ptr("")}); !errors.Is(err, models.ErrInvalidInput) {
		t.Fatalf("want ErrInvalidInput, got %v", err)
	}
	if _, err := svc.UpdateDevice(context.Background(), "d1", models.DeviceUpdate{State: ptr("broken")}); !errors.Is(err, models.ErrInvalidInput) {
		t.Fatalf("want ErrInvalidInput, got %v", err)
	}
}

func TestUpdateDeviceNotFound(t *testing.T) {
	svc := NewDeviceService(newFakeRepo(), &fakeNotifier{})
	if _, err := svc.UpdateDevice(context.Background(), "nope", models.DeviceUpdate{Name: ptr("x")}); !errors.Is(err, models.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestDeleteDeviceInUse(t *testing.T) {
	repo := newFakeRepo(device(models.StateInUse))
	svc := NewDeviceService(repo, &fakeNotifier{})

	if err := svc.DeleteDevice(context.Background(), "d1"); !errors.Is(err, models.ErrDeviceInUse) {
		t.Fatalf("want ErrDeviceInUse, got %v", err)
	}
	if _, ok := repo.devices["d1"]; !ok {
		t.Fatal("device was deleted")
	}
}

func TestDeleteDevice(t *testing.T) {
	repo := newFakeRepo(device(models.StateInactive))
	svc := NewDeviceService(repo, &fakeNotifier{})

	if err := svc.DeleteDevice(context.Background(), "d1"); err != nil {
		t.Fatal(err)
	}
	if _, ok := repo.devices["d1"]; ok {
		t.Fatal("device still exists")
	}
	if err := svc.DeleteDevice(context.Background(), "d1"); !errors.Is(err, models.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}
