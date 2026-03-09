package modbus

import "stellar/internal/telemetry/domain"

type AddressMapper struct{}

func NewAddressMapper() *AddressMapper {
	return &AddressMapper{}
}

func (m *AddressMapper) MappingsFor(_ domain.Asset) []domain.RegisterMapping {
	// TODO: load register mappings from configuration or domain definitions.
	return nil
}
