package influxdb

import "stellar/internal/telemetry/domain"

type Point struct {
	Name  string
	Tags  map[string]string
	Value float64
}

type PointMapper struct{}

func NewPointMapper() *PointMapper {
	return &PointMapper{}
}

func (m *PointMapper) Map(measurement domain.Measurement) Point {
	// TODO: map domain measurements to the real InfluxDB point representation.
	return Point{
		Name: measurement.Name,
		Tags: map[string]string{
			"asset_id": string(measurement.AssetID),
		},
		Value: measurement.Value,
	}
}
