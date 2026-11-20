package repository

const (
	GetDevicesQuery = `
		SELECT id, name, brand, state, created_at
		FROM devices
	`

	GetDeviceByIDQuery = `
		SELECT id, name, brand, state, created_at
		FROM devices
		WHERE id = ?
	`

	CreateDeviceQuery = `
		INSERT INTO devices (id, name, brand, state, created_at)
		VALUES (?, ?, ?, ?, strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
		RETURNING id, name, brand, state, created_at
	`

	UpdateDeviceQuery = `
		UPDATE devices
		SET name = ?, brand = ?, state = ?
		WHERE id = ?
		RETURNING id, name, brand, state, created_at
	`

	DeleteDeviceQuery = `
		DELETE FROM devices
		WHERE id = ?
	`
)
