package notifier

import (
	"encoding/json"
	"net/http"

	"github.com/r3labs/sse/v2"

	"github.com/ytobing/devices-api/internal/models"
)

const stream = "devices"

type Notifier struct {
	server *sse.Server
}

type stateChangeEvent struct {
	DeviceID      string `json:"device_id"`
	PreviousState string `json:"previous_state"`
	NewState      string `json:"new_state"`
}

func New() *Notifier {
	server := sse.New()
	server.AutoReplay = false
	server.CreateStream(stream)
	return &Notifier{server: server}
}

func (n *Notifier) Publish(change models.StateChange) {
	data, err := json.Marshal(stateChangeEvent{
		DeviceID:      change.DeviceID,
		PreviousState: change.PreviousState,
		NewState:      change.NewState,
	})
	if err != nil {
		return
	}
	n.server.Publish(stream, &sse.Event{Event: []byte("state_changed"), Data: data})
}

func (n *Notifier) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	q.Set("stream", stream)
	r.URL.RawQuery = q.Encode()
	n.server.ServeHTTP(w, r)
}

func (n *Notifier) Close() {
	n.server.Close()
}
