package app

import (
	"context"
	"time"

	"stellar/internal/telemetry/app/command"
	"stellar/internal/telemetry/domain"
)

type MeasurementRepository interface {
	SaveBatch(ctx context.Context, measurements []domain.Measurement) error
}

type TelemetrySource interface {
	Collect(ctx context.Context, collectedAt time.Time) ([]domain.Measurement, error)
}

type Application struct {
	Commands Commands
}

type Commands struct {
	CollectTelemetry CollectTelemetryHandler
}

type CollectTelemetryHandler struct {
	repository MeasurementRepository
	source     TelemetrySource
}

func NewApplication(repository MeasurementRepository, source TelemetrySource) Application {
	return Application{
		Commands: Commands{
			CollectTelemetry: CollectTelemetryHandler{
				repository: repository,
				source:     source,
			},
		},
	}
}

func (h CollectTelemetryHandler) Handle(ctx context.Context, cmd command.CollectTelemetry) error {
	measurements, err := h.source.Collect(ctx, cmd.CollectedAt)
	if err != nil {
		return err
	}

	// TODO: add command validation, policy enforcement, orchestration, and richer error handling.
	return h.repository.SaveBatch(ctx, measurements)
}
