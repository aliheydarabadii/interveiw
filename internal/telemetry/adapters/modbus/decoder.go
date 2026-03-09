package modbus

import "stellar/internal/telemetry/domain"

type Decoder struct{}

func NewDecoder() *Decoder {
	return &Decoder{}
}

func (d *Decoder) Decode(_ []byte, _ domain.RegisterMapping) (float64, error) {
	_ = d

	// TODO: decode register payloads into typed measurement values.
	return 0, nil
}
