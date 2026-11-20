package rest_test

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ytobing/devices-api/internal/app"
	"github.com/ytobing/devices-api/internal/handler"
	"github.com/ytobing/devices-api/internal/handler/rest"
	"github.com/ytobing/devices-api/internal/notifier"
	"github.com/ytobing/devices-api/internal/repository"
	"github.com/ytobing/devices-api/internal/service"
)

func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	db, err := app.InitSQLite(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	broker := notifier.New()
	svc := service.NewDeviceService(repository.NewRepository(db), broker)
	srv := httptest.NewServer(rest.NewHandler(svc, broker, slog.New(slog.NewTextHandler(io.Discard, nil))))
	t.Cleanup(func() {
		broker.Close()
		srv.Close()
	})
	return srv
}

func do(t *testing.T, srv *httptest.Server, method, path, body string) (*http.Response, []byte) {
	t.Helper()
	req, err := http.NewRequest(method, srv.URL+path, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp, data
}

func expectStatus(t *testing.T, srv *httptest.Server, method, path, body string, want int) []byte {
	t.Helper()
	resp, data := do(t, srv, method, path, body)
	if resp.StatusCode != want {
		t.Fatalf("%s %s: got %d, want %d (body: %s)", method, path, resp.StatusCode, want, data)
	}
	return data
}

func createDevice(t *testing.T, srv *httptest.Server, body string) handler.DeviceResponse {
	t.Helper()
	resp, data := do(t, srv, http.MethodPost, "/devices", body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create: got %d (body: %s)", resp.StatusCode, data)
	}
	var d handler.DeviceResponse
	if err := json.Unmarshal(data, &d); err != nil {
		t.Fatal(err)
	}
	if resp.Header.Get("Location") != "/devices/"+d.ID {
		t.Fatalf("unexpected Location header: %q", resp.Header.Get("Location"))
	}
	return d
}

func TestDeviceLifecycle(t *testing.T) {
	srv := newServer(t)

	d := createDevice(t, srv, `{"name":"iPhone","brand":"Apple"}`)
	if d.State != "AVAILABLE" || d.CreatedAt == "" {
		t.Fatalf("unexpected device: %+v", d)
	}
	createDevice(t, srv, `{"name":"Pixel","brand":"Google","state":"INACTIVE"}`)

	expectStatus(t, srv, http.MethodGet, "/devices/"+d.ID, "", http.StatusOK)

	var list []handler.DeviceResponse
	json.Unmarshal(expectStatus(t, srv, http.MethodGet, "/devices", "", http.StatusOK), &list)
	if len(list) != 2 {
		t.Fatalf("want 2 devices, got %d", len(list))
	}
	json.Unmarshal(expectStatus(t, srv, http.MethodGet, "/devices?brand=Apple", "", http.StatusOK), &list)
	if len(list) != 1 || list[0].ID != d.ID {
		t.Fatalf("brand filter: got %+v", list)
	}
	json.Unmarshal(expectStatus(t, srv, http.MethodGet, "/devices?state=INACTIVE", "", http.StatusOK), &list)
	if len(list) != 1 || list[0].Brand != "Google" {
		t.Fatalf("state filter: got %+v", list)
	}

	expectStatus(t, srv, http.MethodPatch, "/devices/"+d.ID, `{"state":"IN_USE"}`, http.StatusOK)
	expectStatus(t, srv, http.MethodPatch, "/devices/"+d.ID, `{"name":"iPhone 16"}`, http.StatusConflict)
	expectStatus(t, srv, http.MethodPut, "/devices/"+d.ID, `{"name":"iPhone 16","brand":"Apple","state":"IN_USE"}`, http.StatusConflict)
	expectStatus(t, srv, http.MethodDelete, "/devices/"+d.ID, "", http.StatusConflict)

	var updated handler.DeviceResponse
	json.Unmarshal(expectStatus(t, srv, http.MethodPut, "/devices/"+d.ID, `{"name":"iPhone","brand":"Apple","state":"AVAILABLE"}`, http.StatusOK), &updated)
	json.Unmarshal(expectStatus(t, srv, http.MethodPatch, "/devices/"+d.ID, `{"name":"iPhone 16"}`, http.StatusOK), &updated)
	if updated.Name != "iPhone 16" || updated.CreatedAt != d.CreatedAt {
		t.Fatalf("unexpected device after update: %+v", updated)
	}

	expectStatus(t, srv, http.MethodDelete, "/devices/"+d.ID, "", http.StatusNoContent)
	expectStatus(t, srv, http.MethodGet, "/devices/"+d.ID, "", http.StatusNotFound)
}

func TestValidationErrors(t *testing.T) {
	srv := newServer(t)
	d := createDevice(t, srv, `{"name":"iPhone","brand":"Apple"}`)

	tests := []struct {
		method, path, body string
		want               int
	}{
		{http.MethodPost, "/devices", `{"brand":"Apple"}`, http.StatusBadRequest},
		{http.MethodPost, "/devices", `{"name":"x","brand":"y","state":"broken"}`, http.StatusBadRequest},
		{http.MethodPost, "/devices", `{"name":"x","brand":"y","color":"red"}`, http.StatusBadRequest},
		{http.MethodPost, "/devices", `not json`, http.StatusBadRequest},
		{http.MethodGet, "/devices?state=broken", "", http.StatusBadRequest},
		{http.MethodPatch, "/devices/" + d.ID, `{"created_at":"2020-01-01T00:00:00Z"}`, http.StatusBadRequest},
		{http.MethodPatch, "/devices/" + d.ID, `{}`, http.StatusBadRequest},
		{http.MethodPatch, "/devices/" + d.ID, `{"state":"broken"}`, http.StatusBadRequest},
		{http.MethodPut, "/devices/" + d.ID, `{"name":"x","brand":"y"}`, http.StatusBadRequest},
		{http.MethodGet, "/devices/nope", "", http.StatusNotFound},
		{http.MethodPatch, "/devices/nope", `{"name":"x"}`, http.StatusNotFound},
		{http.MethodPut, "/devices/nope", `{"name":"x","brand":"y","state":"AVAILABLE"}`, http.StatusNotFound},
		{http.MethodDelete, "/devices/nope", "", http.StatusNotFound},
	}
	for _, tc := range tests {
		t.Run(tc.method+" "+tc.path+" "+tc.body, func(t *testing.T) {
			data := expectStatus(t, srv, tc.method, tc.path, tc.body, tc.want)
			var e struct{ Code, Message string }
			if err := json.Unmarshal(data, &e); err != nil || e.Code == "" || e.Message == "" {
				t.Fatalf("want error body, got %s", data)
			}
		})
	}
}

type stateChangeEvent struct {
	DeviceID      string `json:"device_id"`
	PreviousState string `json:"previous_state"`
	NewState      string `json:"new_state"`
}

func TestStateChangeEvents(t *testing.T) {
	srv := newServer(t)
	d := createDevice(t, srv, `{"name":"iPhone","brand":"Apple"}`)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/devices/events", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("unexpected content type %q", ct)
	}

	expectStatus(t, srv, http.MethodPatch, "/devices/"+d.ID, `{"name":"iPhone 16"}`, http.StatusOK)
	expectStatus(t, srv, http.MethodPatch, "/devices/"+d.ID, `{"state":"IN_USE"}`, http.StatusOK)

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		data, ok := strings.CutPrefix(scanner.Text(), "data: ")
		if !ok {
			continue
		}
		var e stateChangeEvent
		if err := json.Unmarshal([]byte(data), &e); err != nil {
			t.Fatal(err)
		}
		want := stateChangeEvent{DeviceID: d.ID, PreviousState: "AVAILABLE", NewState: "IN_USE"}
		if e != want {
			t.Fatalf("got %+v, want %+v", e, want)
		}
		return
	}
	t.Fatalf("no event received: %v", scanner.Err())
}

func TestDocs(t *testing.T) {
	srv := newServer(t)

	resp, body := do(t, srv, http.MethodGet, "/docs", "")
	if resp.StatusCode != http.StatusOK || !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") || !strings.Contains(string(body), "/docs/openapi.yaml") {
		t.Fatalf("unexpected docs page: %d %s", resp.StatusCode, resp.Header.Get("Content-Type"))
	}

	resp, body = do(t, srv, http.MethodGet, "/docs/openapi.yaml", "")
	if resp.StatusCode != http.StatusOK || !strings.HasPrefix(string(body), "openapi: 3.0.3") {
		t.Fatalf("unexpected spec: %d %.40s", resp.StatusCode, body)
	}
}
