package repository_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/ytobing/devices-api/internal/app"
	"github.com/ytobing/devices-api/internal/models"
	"github.com/ytobing/devices-api/internal/repository"
)

func newRepo(t *testing.T) repository.Repository {
	t.Helper()
	db, err := app.InitSQLite(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return repository.NewRepository(db)
}

func create(t *testing.T, repo repository.Repository, id, name, brand, state string) models.Device {
	t.Helper()
	d, err := repo.CreateDevice(context.Background(), models.Device{ID: id, Name: name, Brand: brand, State: state})
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestCreateAndGetDevice(t *testing.T) {
	repo := newRepo(t)
	created := create(t, repo, "d1", "iPhone", "Apple", models.StateAvailable)
	if created.CreatedAt == "" {
		t.Fatal("created_at not set")
	}

	got, err := repo.GetDeviceByID(context.Background(), "d1")
	if err != nil {
		t.Fatal(err)
	}
	if got != created {
		t.Fatalf("got %+v, want %+v", got, created)
	}

	if _, err := repo.GetDeviceByID(context.Background(), "nope"); !errors.Is(err, models.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestGetDevicesFilters(t *testing.T) {
	repo := newRepo(t)
	create(t, repo, "d1", "iPhone", "Apple", models.StateAvailable)
	create(t, repo, "d2", "MacBook", "Apple", models.StateInUse)
	create(t, repo, "d3", "Pixel", "Google", models.StateAvailable)

	tests := []struct {
		brand, state string
		want         int
	}{
		{"", "", 3},
		{"Apple", "", 2},
		{"", models.StateAvailable, 2},
		{"Apple", models.StateAvailable, 1},
		{"Nokia", "", 0},
	}
	for _, tc := range tests {
		got, err := repo.GetDevices(context.Background(), tc.brand, tc.state)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != tc.want {
			t.Errorf("brand=%q state=%q: got %d devices, want %d", tc.brand, tc.state, len(got), tc.want)
		}
	}
}

func TestUpdateDevice(t *testing.T) {
	repo := newRepo(t)
	created := create(t, repo, "d1", "iPhone", "Apple", models.StateAvailable)

	updated, err := repo.UpdateDevice(context.Background(), models.Device{ID: "d1", Name: "iPhone 16", Brand: "Apple", State: models.StateInUse})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "iPhone 16" || updated.State != models.StateInUse || updated.CreatedAt != created.CreatedAt {
		t.Fatalf("unexpected device: %+v", updated)
	}

	if _, err := repo.UpdateDevice(context.Background(), models.Device{ID: "nope", Name: "x", Brand: "y", State: models.StateAvailable}); !errors.Is(err, models.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestDeleteDevice(t *testing.T) {
	repo := newRepo(t)
	create(t, repo, "d1", "iPhone", "Apple", models.StateAvailable)

	if err := repo.DeleteDevice(context.Background(), "d1"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetDeviceByID(context.Background(), "d1"); !errors.Is(err, models.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
	if err := repo.DeleteDevice(context.Background(), "d1"); !errors.Is(err, models.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}
