# Devices API

A REST API to manage devices, written in Go. Data is stored in SQLite, and clients can subscribe to device state changes in real time.

## Features

- Create, fetch, fully update (PUT), partially update (PATCH) and delete devices
- Fetch all devices, or filter by brand and/or state
- Get notified of state changes (device ID, previous state, new state) via Server-Sent Events

### Device

| Field | Notes |
|---|---|
| `id` | UUID, set by the server |
| `name` | required |
| `brand` | required |
| `state` | `AVAILABLE` (default), `IN_USE` or `INACTIVE` |
| `created_at` | set by the server, never changes |

### Rules

- `created_at` (and `id`) cannot be updated.
- `name` and `brand` cannot be updated while the device is `IN_USE`.
- Devices that are `IN_USE` cannot be deleted.
- Notifications are sent only after a state change is saved.

## Running

### With Docker

```sh
docker compose up --build
```

The API runs on http://localhost:8080. Data is kept in the `devices-data` volume.

### Locally

Requires Go 1.23+.

```sh
go run .
```

The database file `devices.db` is created in the current folder. Migrations run automatically on startup.

## Tests

```sh
go test ./...
```

## API

Interactive docs (Swagger UI) are served by the app at http://localhost:8080/docs. The spec itself is [`api/openapi.yaml`](api/openapi.yaml), also available at `/docs/openapi.yaml`.

| Method | Path | Description |
|---|---|---|
| `POST` | `/devices` | Create a device |
| `GET` | `/devices` | Fetch all devices (`?brand=` and `?state=` filters are optional) |
| `GET` | `/devices/{id}` | Fetch one device |
| `PUT` | `/devices/{id}` | Fully update a device (`name`, `brand`, `state` required) |
| `PATCH` | `/devices/{id}` | Partially update a device |
| `DELETE` | `/devices/{id}` | Delete a device |
| `GET` | `/devices/events` | Subscribe to state changes (Server-Sent Events) |
| `GET` | `/docs` | Interactive API docs (Swagger UI) |

Errors return `{"code": "...", "message": "..."}` with status `400` (invalid input), `404` (not found), `409` (device in use) or `500`.

### Examples

```sh
# create
curl -X POST localhost:8080/devices -H 'Content-Type: application/json' \
  -d '{"name":"iPhone 15","brand":"Apple"}'

# fetch all / filter
curl localhost:8080/devices
curl 'localhost:8080/devices?brand=Apple&state=AVAILABLE'

# fetch one
curl localhost:8080/devices/<id>

# full update
curl -X PUT localhost:8080/devices/<id> -H 'Content-Type: application/json' \
  -d '{"name":"iPhone 15 Pro","brand":"Apple","state":"AVAILABLE"}'

# partial update (once IN_USE, name/brand changes and delete return 409)
curl -X PATCH localhost:8080/devices/<id> -H 'Content-Type: application/json' \
  -d '{"state":"IN_USE"}'

# set it back, then delete
curl -X PATCH localhost:8080/devices/<id> -H 'Content-Type: application/json' \
  -d '{"state":"AVAILABLE"}'
curl -X DELETE localhost:8080/devices/<id>
```

### Subscribing to state changes

```sh
curl -N localhost:8080/devices/events
```

Every state change sends an event like:

```
id:
data: {"device_id":"<id>","previous_state":"AVAILABLE","new_state":"IN_USE"}
event: state_changed
```

Only state changes are sent: updating just `name` or `brand` sends no event. Only connected clients receive events, and missed events are not replayed after a disconnect or restart.

## Project structure

```
main.go                  entry point
api/                     OpenAPI spec, embedded in the binary and served at /docs
migrations/              database migrations (goose)
internal/
  app/                   config, database setup, wiring, HTTP server
  models/                Device, states, errors
  repository/            SQL queries (sqlx + SQLite)
  service/               business rules and validation
  notifier/              state change events (r3labs/sse)
  handler/               request/response models
  handler/rest/          HTTP handlers and routes (chi)
```

Requests flow `handler/rest` → `service` → `repository`. Each layer depends only on the one below it, and the service publishes notifications through a small interface, so the notifier can be swapped without touching the business logic.

## Notes and future improvements

- The "in use" checks and the update/delete run as separate queries, so two simultaneous requests could slip past a rule. Wrapping them in a transaction (or a conditional `UPDATE ... WHERE state != 'IN_USE'`) would close that gap.
- Notifications live in memory, so subscribers only receive events from the instance they are connected to. Running several instances would need a shared broker (e.g. Redis pub/sub or NATS).
- The port (`:8080`) and database path are fixed in code; reading them from environment variables would make deployment easier.
- No pagination on `GET /devices`, no authentication, and no HTTP server timeouts yet.
- The `/docs` page loads Swagger UI from a CDN, so it needs internet access.
- SQLite fits a single instance; PostgreSQL would be the next step for higher write concurrency.
