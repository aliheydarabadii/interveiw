package modbus

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"stellar/internal/telemetry/domain"
)

func TestNewSource(t *testing.T) {
	t.Parallel()

	validMapping, err := domain.NewRegisterMapping(domain.HoldingRegister, 40100, 40101, true)
	if err != nil {
		t.Fatalf("expected valid register mapping, got %v", err)
	}

	tests := []struct {
		name    string
		config  Config
		wantErr error
	}{
		{
			name: "empty host rejected",
			config: Config{
				Port:            502,
				UnitID:          1,
				RegisterMapping: validMapping,
			},
			wantErr: ErrEmptyHost,
		},
		{
			name: "zero port rejected",
			config: Config{
				Host:            "127.0.0.1",
				UnitID:          1,
				RegisterMapping: validMapping,
			},
			wantErr: ErrZeroPort,
		},
		{
			name: "zero unit id rejected",
			config: Config{
				Host:            "127.0.0.1",
				Port:            502,
				RegisterMapping: validMapping,
			},
			wantErr: ErrZeroUnitID,
		},
		{
			name: "empty register type rejected",
			config: Config{
				Host:   "127.0.0.1",
				Port:   502,
				UnitID: 1,
				RegisterMapping: domain.RegisterMapping{
					SetpointAddress:    40100,
					ActivePowerAddress: 40101,
					SignedValues:       true,
				},
			},
			wantErr: ErrEmptyRegisterType,
		},
		{
			name: "zero setpoint address rejected",
			config: Config{
				Host:   "127.0.0.1",
				Port:   502,
				UnitID: 1,
				RegisterMapping: domain.RegisterMapping{
					RegisterType:       domain.HoldingRegister,
					ActivePowerAddress: 40101,
					SignedValues:       true,
				},
			},
			wantErr: ErrZeroSetpointAddress,
		},
		{
			name: "zero active power address rejected",
			config: Config{
				Host:   "127.0.0.1",
				Port:   502,
				UnitID: 1,
				RegisterMapping: domain.RegisterMapping{
					RegisterType:    domain.HoldingRegister,
					SetpointAddress: 40100,
					SignedValues:    true,
				},
			},
			wantErr: ErrZeroActivePowerAddress,
		},
		{
			name: "unsupported register type rejected",
			config: Config{
				Host:   "127.0.0.1",
				Port:   502,
				UnitID: 1,
				RegisterMapping: domain.RegisterMapping{
					RegisterType:       domain.RegisterType("input"),
					SetpointAddress:    40100,
					ActivePowerAddress: 40101,
					SignedValues:       true,
				},
			},
			wantErr: ErrUnsupportedRegisterType,
		},
		{
			name: "valid config accepted",
			config: Config{
				Host:            "127.0.0.1",
				Port:            502,
				UnitID:          1,
				RegisterMapping: validMapping,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			source, err := NewSource(tt.config, NewAddressMapper(), NewDecoder())

			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if source == nil {
					t.Fatal("expected source to be created")
				}
				return
			}

			if err == nil {
				t.Fatalf("expected error %v, got nil", tt.wantErr)
			}

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestSourceRead(t *testing.T) {
	t.Parallel()

	source, err := NewSource(DefaultConfig(), NewAddressMapper(), NewDecoder())
	if err != nil {
		t.Fatalf("expected valid source, got %v", err)
	}

	plan, err := source.mapper.Map(source.config.RegisterMapping)
	if err != nil {
		t.Fatalf("expected valid read plan, got %v", err)
	}

	registers := make([]uint16, int(plan.quantity))
	registers[plan.setpointIndex] = 100
	registers[plan.activePowerIndex] = 55

	conn := newFakeConn(buildReadHoldingResponse(1, source.config.UnitID, registers))

	source.dialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		if network != "tcp" {
			t.Fatalf("expected tcp network, got %q", network)
		}
		if address != net.JoinHostPort(source.config.Host, "5020") {
			t.Fatalf("unexpected dial address %q", address)
		}
		return conn, nil
	}

	reading, err := source.Read(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if reading.Setpoint != 100 {
		t.Fatalf("expected setpoint 100, got %v", reading.Setpoint)
	}

	if reading.ActivePower != 55 {
		t.Fatalf("expected active power 55, got %v", reading.ActivePower)
	}

	if !conn.closed {
		t.Fatal("expected connection to be closed")
	}

	if conn.deadline.IsZero() {
		t.Fatal("expected connection deadline to be set")
	}

	if len(conn.written) != 12 {
		t.Fatalf("expected 12-byte modbus request, got %d bytes", len(conn.written))
	}

	if functionCode := conn.written[7]; functionCode != readHoldingRegisters {
		t.Fatalf("expected function code %d, got %d", readHoldingRegisters, functionCode)
	}

	if quantity := binary.BigEndian.Uint16(conn.written[10:12]); quantity != plan.quantity {
		t.Fatalf("expected quantity %d, got %d", plan.quantity, quantity)
	}
}

func TestSourceReadUnexpectedTransactionID(t *testing.T) {
	t.Parallel()

	source, err := NewSource(DefaultConfig(), NewAddressMapper(), NewDecoder())
	if err != nil {
		t.Fatalf("expected valid source, got %v", err)
	}

	plan, err := source.mapper.Map(source.config.RegisterMapping)
	if err != nil {
		t.Fatalf("expected valid read plan, got %v", err)
	}

	registers := make([]uint16, int(plan.quantity))
	conn := newFakeConn(buildReadHoldingResponse(99, source.config.UnitID, registers))

	source.dialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		return conn, nil
	}

	_, err = source.Read(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "unexpected transaction id") {
		t.Fatalf("expected unexpected transaction id error, got %v", err)
	}
}

func TestDeadlineFromContextUsesContextDeadline(t *testing.T) {
	t.Parallel()

	expected := time.Now().Add(2 * time.Second).Round(0)
	ctx, cancel := context.WithDeadline(context.Background(), expected)
	defer cancel()

	got := deadlineFromContext(ctx, defaultIOTimeout)
	if !got.Equal(expected) {
		t.Fatalf("expected deadline %v, got %v", expected, got)
	}
}

func TestDeadlineFromContextUsesFallback(t *testing.T) {
	t.Parallel()

	before := time.Now()
	got := deadlineFromContext(context.Background(), defaultIOTimeout)
	after := time.Now()

	min := before.Add(defaultIOTimeout)
	max := after.Add(defaultIOTimeout + 100*time.Millisecond)

	if got.Before(min) || got.After(max) {
		t.Fatalf("expected deadline between %v and %v, got %v", min, max, got)
	}
}

func TestWriteAllWritesEntirePayload(t *testing.T) {
	t.Parallel()

	conn := &fakeConn{writeChunkSize: 3}
	payload := []byte{1, 2, 3, 4, 5, 6, 7}

	if err := writeAll(conn, payload); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !bytes.Equal(conn.written, payload) {
		t.Fatalf("expected payload %v, got %v", payload, conn.written)
	}
}

func TestNextTransactionIDWrapsToOne(t *testing.T) {
	t.Parallel()

	source, err := NewSource(DefaultConfig(), NewAddressMapper(), NewDecoder())
	if err != nil {
		t.Fatalf("expected valid source, got %v", err)
	}

	source.transactionID = ^uint16(0)

	got := source.nextTransactionID()
	if got != 1 {
		t.Fatalf("expected wrapped transaction id 1, got %d", got)
	}
}

func buildReadHoldingResponse(transactionID uint16, unitID uint8, registers []uint16) []byte {
	byteCount := len(registers) * 2
	length := uint16(3 + byteCount)

	response := make([]byte, 7+2+byteCount)
	binary.BigEndian.PutUint16(response[0:2], transactionID)
	binary.BigEndian.PutUint16(response[2:4], 0)
	binary.BigEndian.PutUint16(response[4:6], length)
	response[6] = unitID
	response[7] = readHoldingRegisters
	response[8] = byte(byteCount)

	offset := 9
	for _, register := range registers {
		binary.BigEndian.PutUint16(response[offset:offset+2], register)
		offset += 2
	}

	return response
}

type fakeConn struct {
	reader         *bytes.Reader
	written        []byte
	writeChunkSize int
	deadline       time.Time
	closed         bool
}

func newFakeConn(response []byte) *fakeConn {
	return &fakeConn{
		reader: bytes.NewReader(response),
	}
}

func (c *fakeConn) Read(p []byte) (int, error) {
	return c.reader.Read(p)
}

func (c *fakeConn) Write(p []byte) (int, error) {
	n := len(p)
	if c.writeChunkSize > 0 && n > c.writeChunkSize {
		n = c.writeChunkSize
	}

	c.written = append(c.written, p[:n]...)
	return n, nil
}

func (c *fakeConn) Close() error {
	c.closed = true
	return nil
}

func (c *fakeConn) LocalAddr() net.Addr {
	return fakeAddr("local")
}

func (c *fakeConn) RemoteAddr() net.Addr {
	return fakeAddr("remote")
}

func (c *fakeConn) SetDeadline(t time.Time) error {
	c.deadline = t
	return nil
}

func (c *fakeConn) SetReadDeadline(t time.Time) error {
	c.deadline = t
	return nil
}

func (c *fakeConn) SetWriteDeadline(t time.Time) error {
	c.deadline = t
	return nil
}

type fakeAddr string

func (a fakeAddr) Network() string { return "tcp" }
func (a fakeAddr) String() string  { return string(a) }

var _ net.Conn = (*fakeConn)(nil)
var _ io.Reader = (*bytes.Reader)(nil)
