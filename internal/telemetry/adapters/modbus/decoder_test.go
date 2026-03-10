package modbus

import "testing"

func TestDecoderDecodeRegister(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		raw          uint16
		signedValues bool
		want         float64
	}{
		{
			name:         "signed register is decoded as negative int16",
			raw:          0xFFFF,
			signedValues: true,
			want:         -1,
		},
		{
			name:         "unsigned register keeps raw value",
			raw:          0xFFFF,
			signedValues: false,
			want:         65535,
		},
	}

	decoder := NewDecoder()

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := decoder.DecodeRegister(tt.raw, tt.signedValues); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
