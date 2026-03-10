# Integration Service

The Integration Service is the command-side telemetry ingester for the backend system.

It polls Modbus TCP registers, converts raw values into telemetry readings, validates domain measurements, and writes valid measurements to InfluxDB.

## Purpose

- collect telemetry for a configured asset
- validate telemetry using domain rules before persistence
- persist valid measurements to InfluxDB
- expose minimal health and readiness endpoints for runtime supervision

## Architecture

This service follows a lightweight DDD + CQRS + Clean Architecture style:

- `domain`: business concepts and validation rules
- `app`: command-side orchestration
- `adapters`: infrastructure implementations for Modbus and InfluxDB
- `ports`: runtime concerns such as worker execution and HTTP endpoints
- `cmd/integration-service`: bootstrap and wiring

CQRS note:

- this service is command-side only
- there are no query handlers or read-side endpoints in this service

## Package Structure

```text
cmd/
  integration-service/
    main.go

internal/
  telemetry/
    app/
      app.go
      command/
        collect_telemetry.go

    domain/
      asset.go
      register_mapping.go
      measurement.go
      errors.go

    ports/
      worker.go
      http.go

    adapters/
      modbus/
        source.go
        decoder.go
        address_mapper.go
      influxdb/
        measurement_repository.go
        point_mapper.go
```

## Runtime Flow

1. A ticker triggers telemetry collection on the configured poll interval.
2. The worker builds `CollectTelemetry{CollectedAt: time.Now().UTC()}`.
3. The command handler reads raw telemetry from the Modbus adapter.
4. The application constructs a domain `Measurement`.
5. Invalid measurements are rejected and skipped.
6. Valid measurements are persisted through the InfluxDB adapter.

## Assumptions

- time is passed through the command via `CollectedAt`
- register mapping is configuration-driven
- invalid measurements are skipped and not persisted
- the default Modbus mapping uses holding registers `40100` and `40101`
- when `signed_values=true`, Modbus registers are decoded as signed 16-bit integers

## Configuration

The service loads configuration from environment variables.

Required variables:

```text
ASSET_ID
ASSET_TYPE
MODBUS_HOST
MODBUS_PORT
MODBUS_UNIT_ID
MODBUS_REGISTER_TYPE
MODBUS_SETPOINT_ADDRESS
MODBUS_ACTIVE_POWER_ADDRESS
MODBUS_SIGNED_VALUES
POLL_INTERVAL
INFLUX_URL
INFLUX_TOKEN
INFLUX_ORG
INFLUX_BUCKET
HTTP_PORT
```

Example:

```bash
export ASSET_ID=871689260010377213
export ASSET_TYPE=solar_panel
export MODBUS_HOST=127.0.0.1
export MODBUS_PORT=5020
export MODBUS_UNIT_ID=1
export MODBUS_REGISTER_TYPE=holding
export MODBUS_SETPOINT_ADDRESS=40100
export MODBUS_ACTIVE_POWER_ADDRESS=40101
export MODBUS_SIGNED_VALUES=true
export POLL_INTERVAL=1s
export INFLUX_URL=http://127.0.0.1:8086
export INFLUX_TOKEN=dev-token
export INFLUX_ORG=local
export INFLUX_BUCKET=telemetry
export HTTP_PORT=8080
```

## Run

```bash
go run ./cmd/integration-service
```

Endpoints:

- `GET /healthz`
- `GET /readyz`

## Docker Compose

To run the Integration Service with InfluxDB and a local Modbus simulator:

```bash
docker compose up --build
```

Services started by the compose stack:

- `integration-service`
- `influxdb`
- `modbus-simulator`

The compose stack configures the simulator via `MODBUS_SERVER_CONFIG`, including the holding register values for `40100` and `40101`.

Exposed ports:

- `8080`: Integration Service HTTP endpoints
- `8086`: InfluxDB
- `5020`: Modbus simulator

## Testing

Run the full automated test suite with:

```bash
go test ./...
```

The test suite covers domain validation, application command orchestration, Modbus adapter behavior, InfluxDB point mapping/repository behavior, and worker loop behavior.
