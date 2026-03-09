package influxdb

import (
	"time"

	"stellar/internal/telemetry/domain"
)

type Point struct {
	Name      string
	Tags      map[string]string
	Fields    map[string]float64
	Timestamp time.Time
}

type PointMapper struct{}

func NewPointMapper() *PointMapper {
	return &PointMapper{}
}

func (m *PointMapper) Map(measurement domain.Measurement) Point {
	// TODO: map domain measurements to the real InfluxDB point representation.
	return Point{
		Name: "telemetry",
		Tags: map[string]string{
			"asset_id": measurement.AssetID.String(),
		},
		Fields: map[string]float64{
			"setpoint":     measurement.Setpoint,
			"active_power": measurement.ActivePower,
		},
		Timestamp: measurement.CollectedAt,
	}
}
