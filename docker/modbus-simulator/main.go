package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	readHoldingRegisters = 0x03
	modbusExceptionFC    = 0x80
)

type config struct {
	listenerAddress string
	port            uint16
	unitID          uint8
	registers       map[uint16]uint16
}

type serverConfigEnvelope struct {
	Server struct {
		ListenerAddress string `json:"listenerAddress"`
		ListenerPort    uint16 `json:"listenerPort"`
		Protocol        string `json:"protocol"`
	} `json:"server"`
	Registers struct {
		HoldingRegister map[string]string `json:"holdingRegister"`
	} `json:"registers"`
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, cfg); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("modbus simulator stopped with error: %v", err)
	}
}

func run(ctx context.Context, cfg config) error {
	listener, err := net.Listen("tcp", net.JoinHostPort(cfg.listenerAddress, strconv.Itoa(int(cfg.port))))
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	defer listener.Close()

	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()

	log.Printf("modbus simulator listening on %s:%d", cfg.listenerAddress, cfg.port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("accept: %w", err)
		}

		go handleConn(ctx, conn, cfg)
	}
}

func handleConn(ctx context.Context, conn net.Conn, cfg config) {
	defer conn.Close()

	for {
		if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
			return
		}

		request := make([]byte, 12)
		if _, err := io.ReadFull(conn, request); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
				return
			}
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		response, err := buildResponse(request, cfg)
		if err != nil {
			response = buildExceptionResponse(request, cfg.unitID, 0x01)
		}

		if _, err := conn.Write(response); err != nil {
			return
		}
	}
}

func buildResponse(request []byte, cfg config) ([]byte, error) {
	if len(request) != 12 {
		return nil, fmt.Errorf("invalid request length")
	}

	transactionID := binary.BigEndian.Uint16(request[0:2])
	protocolID := binary.BigEndian.Uint16(request[2:4])
	if protocolID != 0 {
		return nil, fmt.Errorf("unsupported protocol id")
	}

	unitID := request[6]
	if unitID != cfg.unitID {
		return nil, fmt.Errorf("unexpected unit id")
	}

	functionCode := request[7]
	if functionCode != readHoldingRegisters {
		return nil, fmt.Errorf("unsupported function code")
	}

	startAddress := binary.BigEndian.Uint16(request[8:10])
	quantity := binary.BigEndian.Uint16(request[10:12])
	if quantity == 0 {
		return nil, fmt.Errorf("invalid quantity")
	}

	registers := make([]uint16, quantity)
	for i := uint16(0); i < quantity; i++ {
		registers[i] = registerValue(startAddress+i, cfg)
	}

	byteCount := len(registers) * 2
	response := make([]byte, 9+byteCount)
	binary.BigEndian.PutUint16(response[0:2], transactionID)
	binary.BigEndian.PutUint16(response[2:4], 0)
	binary.BigEndian.PutUint16(response[4:6], uint16(3+byteCount))
	response[6] = cfg.unitID
	response[7] = readHoldingRegisters
	response[8] = byte(byteCount)

	offset := 9
	for _, register := range registers {
		binary.BigEndian.PutUint16(response[offset:offset+2], register)
		offset += 2
	}

	return response, nil
}

func buildExceptionResponse(request []byte, unitID uint8, exceptionCode byte) []byte {
	transactionID := binary.BigEndian.Uint16(request[0:2])
	functionCode := byte(readHoldingRegisters)
	if len(request) >= 8 {
		functionCode = request[7]
	}

	response := make([]byte, 9)
	binary.BigEndian.PutUint16(response[0:2], transactionID)
	binary.BigEndian.PutUint16(response[2:4], 0)
	binary.BigEndian.PutUint16(response[4:6], 3)
	response[6] = unitID
	response[7] = functionCode | byte(modbusExceptionFC)
	response[8] = exceptionCode

	return response
}

func registerValue(offset uint16, cfg config) uint16 {
	return cfg.registers[offset]
}

func loadConfig() (config, error) {
	if raw := os.Getenv("MODBUS_SERVER_CONFIG"); raw != "" {
		return loadJSONConfig(raw)
	}

	port, err := requiredUint16("MODBUS_SIM_PORT")
	if err != nil {
		return config{}, err
	}

	unitID, err := requiredUint8("MODBUS_SIM_UNIT_ID")
	if err != nil {
		return config{}, err
	}

	setpoint, err := requiredInt16("MODBUS_SIM_SETPOINT")
	if err != nil {
		return config{}, err
	}

	activePower, err := requiredInt16("MODBUS_SIM_ACTIVE_POWER")
	if err != nil {
		return config{}, err
	}

	return config{
		listenerAddress: "",
		port:            port,
		unitID:          unitID,
		registers: map[uint16]uint16{
			99:  uint16(setpoint),
			100: uint16(activePower),
		},
	}, nil
}

func loadJSONConfig(raw string) (config, error) {
	var envelope serverConfigEnvelope
	if err := json.Unmarshal([]byte(raw), &envelope); err != nil {
		return config{}, fmt.Errorf("parse MODBUS_SERVER_CONFIG: %w", err)
	}

	if envelope.Server.Protocol != "" && !strings.EqualFold(envelope.Server.Protocol, "TCP") {
		return config{}, fmt.Errorf("unsupported protocol %q", envelope.Server.Protocol)
	}

	if envelope.Server.ListenerPort == 0 {
		return config{}, fmt.Errorf("missing listener port")
	}

	registers, err := parseHoldingRegisters(envelope.Registers.HoldingRegister)
	if err != nil {
		return config{}, err
	}

	return config{
		listenerAddress: envelope.Server.ListenerAddress,
		port:            envelope.Server.ListenerPort,
		unitID:          1,
		registers:       registers,
	}, nil
}

func parseHoldingRegisters(values map[string]string) (map[uint16]uint16, error) {
	registers := make(map[uint16]uint16, len(values))

	for addressText, rawValue := range values {
		address, err := strconv.ParseUint(addressText, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("parse holding register address %q: %w", addressText, err)
		}

		if address < 40001 {
			return nil, fmt.Errorf("holding register address %d is below base 40001", address)
		}

		value, err := strconv.ParseUint(rawValue, 0, 16)
		if err != nil {
			return nil, fmt.Errorf("parse holding register value %q: %w", rawValue, err)
		}

		registers[uint16(address-40001)] = uint16(value)
	}

	return registers, nil
}

func requiredUint16(key string) (uint16, error) {
	value := os.Getenv(key)
	if value == "" {
		return 0, fmt.Errorf("missing %s", key)
	}

	parsed, err := strconv.ParseUint(value, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}

	return uint16(parsed), nil
}

func requiredUint8(key string) (uint8, error) {
	value := os.Getenv(key)
	if value == "" {
		return 0, fmt.Errorf("missing %s", key)
	}

	parsed, err := strconv.ParseUint(value, 10, 8)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}

	return uint8(parsed), nil
}

func requiredInt16(key string) (int16, error) {
	value := os.Getenv(key)
	if value == "" {
		return 0, fmt.Errorf("missing %s", key)
	}

	parsed, err := strconv.ParseInt(value, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}

	return int16(parsed), nil
}
