package repository

const (
	GetDevicesQuery = `
		SELECT id, name, brand, state, created_at
		FROM devices
	`
)
