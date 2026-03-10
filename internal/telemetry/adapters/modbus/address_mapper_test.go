package modbus

import (
	"testing"

	"stellar/internal/telemetry/domain"
)

func TestAddressMapperMap(t *testing.T) {
	t.Parallel()

	mapping, err := domain.NewRegisterMapping(domain.HoldingRegister, 40100, 40101, true)
	if err != nil {
		t.Fatalf("expected valid register mapping, got %v", err)
	}

	mapper := NewAddressMapper()
	plan, err := mapper.Map(mapping)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if plan.startAddress != 99 {
		t.Fatalf("expected start address %d, got %d", 99, plan.startAddress)
	}

	if plan.quantity != 2 {
		t.Fatalf("expected quantity %d, got %d", 2, plan.quantity)
	}

	if plan.setpointIndex != 0 {
		t.Fatalf("expected setpoint index %d, got %d", 0, plan.setpointIndex)
	}

	if plan.activePowerIndex != 1 {
		t.Fatalf("expected active power index %d, got %d", 1, plan.activePowerIndex)
	}

	if !plan.signedValues {
		t.Fatal("expected signed values to be true")
	}
}
