package domain

import "testing"

func TestNewRegisterMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		registerType       RegisterType
		setpointAddress    uint16
		activePowerAddress uint16
		signedValues       bool
		wantErrText        string
	}{
		{
			name:               "empty register type rejected",
			registerType:       "",
			setpointAddress:    40100,
			activePowerAddress: 40101,
			signedValues:       true,
			wantErrText:        "register type must not be empty",
		},
		{
			name:               "zero setpoint address rejected",
			registerType:       HoldingRegister,
			setpointAddress:    0,
			activePowerAddress: 40101,
			signedValues:       true,
			wantErrText:        "setpoint address must not be zero",
		},
		{
			name:               "zero active power address rejected",
			registerType:       HoldingRegister,
			setpointAddress:    40100,
			activePowerAddress: 0,
			signedValues:       true,
			wantErrText:        "active power address must not be zero",
		},
		{
			name:               "equal addresses rejected",
			registerType:       HoldingRegister,
			setpointAddress:    40100,
			activePowerAddress: 40100,
			signedValues:       true,
			wantErrText:        "setpoint and active power addresses must be different",
		},
		{
			name:               "valid mapping accepted",
			registerType:       HoldingRegister,
			setpointAddress:    40100,
			activePowerAddress: 40101,
			signedValues:       true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mapping, err := NewRegisterMapping(tt.registerType, tt.setpointAddress, tt.activePowerAddress, tt.signedValues)
			if tt.wantErrText == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}

				if mapping.RegisterType != tt.registerType {
					t.Fatalf("expected register type %q, got %q", tt.registerType, mapping.RegisterType)
				}

				if mapping.SetpointAddress != tt.setpointAddress {
					t.Fatalf("expected setpoint address %d, got %d", tt.setpointAddress, mapping.SetpointAddress)
				}

				if mapping.ActivePowerAddress != tt.activePowerAddress {
					t.Fatalf("expected active power address %d, got %d", tt.activePowerAddress, mapping.ActivePowerAddress)
				}

				if mapping.SignedValues != tt.signedValues {
					t.Fatalf("expected signed values %t, got %t", tt.signedValues, mapping.SignedValues)
				}

				return
			}

			if err == nil {
				t.Fatalf("expected error %q, got nil", tt.wantErrText)
			}

			if err.Error() != tt.wantErrText {
				t.Fatalf("expected error %q, got %q", tt.wantErrText, err.Error())
			}
		})
	}
}

func TestNewDefaultRegisterMapping(t *testing.T) {
	t.Parallel()

	mapping := NewDefaultRegisterMapping()

	if mapping.RegisterType != HoldingRegister {
		t.Fatalf("expected register type %q, got %q", HoldingRegister, mapping.RegisterType)
	}

	if mapping.SetpointAddress != 40100 {
		t.Fatalf("expected setpoint address %d, got %d", 40100, mapping.SetpointAddress)
	}

	if mapping.ActivePowerAddress != 40101 {
		t.Fatalf("expected active power address %d, got %d", 40101, mapping.ActivePowerAddress)
	}

	if !mapping.SignedValues {
		t.Fatal("expected signed values to be true")
	}
}
