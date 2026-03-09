package modbus

import (
	"context"
	"time"

	"stellar/internal/telemetry/domain"
)

type Source struct {
	mapper  *AddressMapper
	decoder *Decoder
}

func NewSource(mapper *AddressMapper, decoder *Decoder) *Source {
	return &Source{
		mapper:  mapper,
		decoder: decoder,
	}
}

func (s *Source) Collect(_ context.Context, collectedAt time.Time) ([]domain.Measurement, error) {
	_ = s.mapper
	_ = s.decoder

	measurement, err := domain.NewMeasurement(domain.DefaultAssetID, 0, 0, collectedAt)
	if err != nil {
		return nil, err
	}

	// TODO: replace with real Modbus polling, decoding, and domain mapping.
	return []domain.Measurement{
		measurement,
	}, nil
}
