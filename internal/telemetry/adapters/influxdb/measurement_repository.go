package influxdb

import (
	"context"

	"stellar/internal/telemetry/domain"
)

type MeasurementRepository struct {
	mapper *PointMapper
}

func NewMeasurementRepository(mapper *PointMapper) *MeasurementRepository {
	return &MeasurementRepository{
		mapper: mapper,
	}
}

func (r *MeasurementRepository) SaveBatch(_ context.Context, measurements []domain.Measurement) error {
	for _, measurement := range measurements {
		_ = r.mapper.Map(measurement)
	}

	// TODO: replace with real InfluxDB batch persistence.
	return nil
}
