package domain

type MeasurementPolicy struct{}

func (p MeasurementPolicy) Allows(_ Measurement) bool {
	// TODO: apply telemetry validation and filtering rules.
	return true
}
