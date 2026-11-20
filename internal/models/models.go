package models

import "errors"

var (
	ErrNotFound     = errors.New("device not found")
	ErrDeviceInUse  = errors.New("device is in use")
	ErrInvalidInput = errors.New("invalid input")
)

type Device struct {
	ID        string `db:"id"`
	Name      string `db:"name"`
	Brand     string `db:"brand"`
	State     string `db:"state"`
	CreatedAt string `db:"created_at"`
}

type DeviceUpdate struct {
	Name  *string
	Brand *string
	State *string
}

type StateChange struct {
	DeviceID      string
	PreviousState string
	NewState      string
}

const (
	StateAvailable = "AVAILABLE"
	StateInUse     = "IN_USE"
	StateInactive  = "INACTIVE"
)

func IsValidState(s string) bool {
	switch s {
	case StateAvailable, StateInUse, StateInactive:
		return true
	}
	return false
}
