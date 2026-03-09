package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewMeasurement(t *testing.T) {
	t.Parallel()

	collectedAt := time.Date(2026, time.March, 9, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		setpoint    float64
		activePower float64
		wantErr     error
	}{
		{
			name:        "negative setpoint rejected",
			setpoint:    -1,
			activePower: 0,
			wantErr:     ErrNegativeSetpoint,
		},
		{
			name:        "negative active power rejected",
			setpoint:    10,
			activePower: -1,
			wantErr:     ErrNegativeActivePower,
		},
		{
			name:        "active power greater than setpoint rejected",
			setpoint:    10,
			activePower: 11,
			wantErr:     ErrActivePowerExceedsSetpoint,
		},
		{
			name:        "valid measurement accepted",
			setpoint:    10,
			activePower: 8,
			wantErr:     nil,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			measurement, err := NewMeasurement(DefaultAssetID, tt.setpoint, tt.activePower, collectedAt)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}

				if measurement.AssetID != DefaultAssetID {
					t.Fatalf("expected asset ID %q, got %q", DefaultAssetID, measurement.AssetID)
				}

				if measurement.Setpoint != tt.setpoint {
					t.Fatalf("expected setpoint %v, got %v", tt.setpoint, measurement.Setpoint)
				}

				if measurement.ActivePower != tt.activePower {
					t.Fatalf("expected active power %v, got %v", tt.activePower, measurement.ActivePower)
				}

				if !measurement.CollectedAt.Equal(collectedAt) {
					t.Fatalf("expected collected at %v, got %v", collectedAt, measurement.CollectedAt)
				}

				return
			}

			if err == nil {
				t.Fatalf("expected error %v, got nil", tt.wantErr)
			}

			if !errors.Is(err, ErrInvalidMeasurement) {
				t.Fatalf("expected invalid measurement error, got %v", err)
			}

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}
