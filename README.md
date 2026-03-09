# Integration Service

This repository currently contains only the initial scaffold for the Integration Service microservice.

Scope of this step:

- minimal Go project structure
- DDD-lite + CQRS + Clean Architecture oriented package layout
- placeholder domain, application, ports, and adapters
- bootstrap wiring with a health endpoint, background worker loop, and graceful shutdown

Not implemented yet:

- real Modbus integration
- real InfluxDB persistence
- business rules and telemetry collection flow
- full test suite

## Run

```bash
go run ./cmd/integration-service
```

The service exposes a simple health endpoint at `GET /healthz` on port `8080`.
