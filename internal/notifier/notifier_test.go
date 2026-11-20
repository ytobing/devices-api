package notifier

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ytobing/devices-api/internal/models"
)

func subscribe(t *testing.T, n *Notifier) *bufio.Scanner {
	t.Helper()
	srv := httptest.NewServer(n)
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("unexpected content type %q", ct)
	}
	return bufio.NewScanner(resp.Body)
}

func TestPublishSendsStateChangeEvent(t *testing.T) {
	n := New()
	t.Cleanup(n.Close)
	scanner := subscribe(t, n)

	n.Publish(models.StateChange{DeviceID: "d1", PreviousState: models.StateAvailable, NewState: models.StateInUse})

	var event, data string
	for scanner.Scan() {
		line := scanner.Text()
		if v, ok := strings.CutPrefix(line, "event: "); ok {
			event = v
		}
		if v, ok := strings.CutPrefix(line, "data: "); ok {
			data = v
		}
		if line == "" && data != "" {
			break
		}
	}

	if event != "state_changed" {
		t.Fatalf("got event %q", event)
	}
	var got stateChangeEvent
	if err := json.Unmarshal([]byte(data), &got); err != nil {
		t.Fatal(err)
	}
	want := stateChangeEvent{DeviceID: "d1", PreviousState: models.StateAvailable, NewState: models.StateInUse}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestCloseEndsSubscriptions(t *testing.T) {
	n := New()
	scanner := subscribe(t, n)

	n.Close()

	done := make(chan struct{})
	go func() {
		for scanner.Scan() {
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("stream still open after Close")
	}
}
