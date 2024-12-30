package rtu

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/rinzlerlabs/gomodbus/client"
	"github.com/rinzlerlabs/gomodbus/client/serial"
	"github.com/rinzlerlabs/gomodbus/common"
	"github.com/rinzlerlabs/gomodbus/transport"
	st "github.com/rinzlerlabs/gomodbus/transport/serial"
	"github.com/rinzlerlabs/gomodbus/transport/serial/rtu"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func newModbusClient(logger *zap.Logger, stream io.ReadWriteCloser, responseTimeout time.Duration) client.ModbusClient {
	ctx := context.Background()
	t := rtu.NewModbusClientTransport(stream, logger)
	newHeader := func(address uint16) transport.Header {
		return st.NewHeader(address)
	}
	requestCreator := serial.NewSerialRequestCreator(newHeader, rtu.NewModbusRequest)
	return client.NewModbusClient(ctx, logger, t, requestCreator, responseTimeout)
}

type testSerialPort struct {
	readData  []byte
	writeData []byte
}

func (t *testSerialPort) Read(b []byte) (n int, err error) {
	if len(t.readData) == 0 {
		return 0, io.EOF
	}
	lenRead := copy(b, t.readData)
	t.readData = t.readData[lenRead:]
	return lenRead, nil
}

func (t *testSerialPort) Write(b []byte) (n int, err error) {
	t.writeData = b
	return len(b), nil
}

func (t *testSerialPort) Close() error {
	return nil
}

func TestRTUReadCoils(t *testing.T) {
	tests := []struct {
		name            string
		toServer        []byte
		coils           []bool
		fromServerError error
		fromServer      []byte
	}{
		{
			name:       "Valid",
			toServer:   []byte{0x04, 0x01, 0x00, 0x0A, 0x00, 0x0D, 0xDD, 0x98},
			coils:      []bool{false, true, false, true, false, false, false, false, true, false, false, false, true},
			fromServer: []byte{0x04, 0x01, 0x02, 0x0A, 0x11, 0xB3, 0x50},
		},
		{
			name:            "ServerError_IllegalDataAddress",
			toServer:        []byte{0x04, 0x01, 0x00, 0x0A, 0x00, 0x0D, 0xDD, 0x98},
			fromServer:      []byte{0x04, 0x81, 0x02, 0xD1, 0x90},
			fromServerError: common.ErrIllegalDataAddress,
		},
		{
			name:            "InvalidRequest_IvalidChecksum",
			toServer:        []byte{0x04, 0x01, 0x00, 0x0A, 0x00, 0x0D, 0xDD, 0x98},
			fromServerError: common.ErrInvalidChecksum,
			fromServer:      []byte{0x04, 0x01, 0x02, 0x0A, 0x11, 0xB3, 0x51},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			port := &testSerialPort{
				readData: []byte(tt.fromServer),
			}
			client := newModbusClient(logger, port, 1*time.Minute)
			resp, err := client.ReadCoils(0x04, 10, 13)
			if tt.fromServerError != nil {
				assert.Equal(t, tt.fromServerError, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.toServer, port.writeData)
			assert.Equal(t, tt.coils, resp)
		})
	}
}

func TestRTUReadDiscreteInputs(t *testing.T) {
	tests := []struct {
		name            string
		toServer        []byte
		inputs          []bool
		fromServerError error
		fromServer      []byte
	}{
		{
			name:       "Valid",
			toServer:   []byte{0x04, 0x02, 0x00, 0x0A, 0x00, 0x0D, 0x99, 0x98},
			inputs:     []bool{false, true, false, true, false, false, false, false, true, false, false, false, true},
			fromServer: []byte{0x04, 0x02, 0x02, 0x0A, 0x11, 0xb3, 0x14},
		},
		{
			name:            "ServerError_IllegalDataAddress",
			toServer:        []byte{0x04, 0x02, 0x00, 0x0A, 0x00, 0x0D, 0x99, 0x98},
			fromServer:      []byte{0x04, 0x82, 0x02, 0xD1, 0x60},
			fromServerError: common.ErrIllegalDataAddress,
		},
		{
			name:            "InvalidRequest_InvalidChecksum",
			toServer:        []byte{0x04, 0x02, 0x00, 0x0A, 0x00, 0x0D, 0x99, 0x98},
			fromServerError: common.ErrInvalidChecksum,
			fromServer:      []byte{0x04, 0x02, 0x02, 0x0A, 0x11, 0xB3, 0x15},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			port := &testSerialPort{
				readData: []byte(tt.fromServer),
			}
			client := newModbusClient(logger, port, 1*time.Minute)
			resp, err := client.ReadDiscreteInputs(0x04, 10, 13)
			if tt.fromServerError != nil {
				assert.Equal(t, tt.fromServerError, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.toServer, port.writeData)
			assert.Equal(t, tt.inputs, resp)
		})
	}
}

func TestRTUReadHoldingRegisters(t *testing.T) {
	tests := []struct {
		name            string
		toServer        []byte
		registers       []uint16
		fromServerError error
		fromServer      []byte
	}{
		{
			name:       "Valid",
			toServer:   []byte{0x04, 0x03, 0x00, 0x00, 0x00, 0x02, 0xc4, 0x5e},
			registers:  []uint16{0x0006, 0x0005},
			fromServer: []byte{0x04, 0x03, 0x04, 0x00, 0x06, 0x00, 0x05, 0x8f, 0x31},
		},
		{
			name:            "ServerError_IllegalDataAddress",
			toServer:        []byte{0x04, 0x03, 0x00, 0x00, 0x00, 0x02, 0xc4, 0x5e},
			fromServer:      []byte{0x04, 0x83, 0x02, 0xD0, 0xF0},
			fromServerError: common.ErrIllegalDataAddress,
		},
		{
			name:            "InvalidRequest_InvalidChecksum",
			toServer:        []byte{0x04, 0x03, 0x00, 0x00, 0x00, 0x02, 0xc4, 0x5e},
			fromServerError: common.ErrInvalidChecksum,
			fromServer:      []byte{0x04, 0x03, 0x04, 0x00, 0x06, 0x00, 0x05, 0x8f, 0x32},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			port := &testSerialPort{
				readData: []byte(tt.fromServer),
			}
			client := newModbusClient(logger, port, 1*time.Minute)
			resp, err := client.ReadHoldingRegisters(0x04, 0, 2)
			if tt.fromServerError != nil {
				assert.Equal(t, tt.fromServerError, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.toServer, port.writeData)
			assert.Equal(t, tt.registers, resp)
		})
	}
}

func TestRTUReadInputRegisters(t *testing.T) {
	tests := []struct {
		name            string
		toServer        []byte
		registers       []uint16
		fromServerError error
		fromServer      []byte
	}{
		{
			name:       "Valid",
			toServer:   []byte{0x04, 0x04, 0x00, 0x00, 0x00, 0x02, 0x71, 0x9e},
			registers:  []uint16{0x0006, 0x0005},
			fromServer: []byte{0x04, 0x04, 0x04, 0x00, 0x06, 0x00, 0x05, 0x8e, 0x86},
		},
		{
			name:            "ServerError_IllegalDataAddress",
			toServer:        []byte{0x04, 0x04, 0x00, 0x00, 0x00, 0x02, 0x71, 0x9e},
			fromServer:      []byte{0x04, 0x84, 0x02, 0xD2, 0xC0},
			fromServerError: common.ErrIllegalDataAddress,
		},
		{
			name:            "InvalidRequest_InvalidChecksum",
			toServer:        []byte{0x04, 0x04, 0x00, 0x00, 0x00, 0x02, 0x71, 0x9e},
			fromServerError: common.ErrInvalidChecksum,
			fromServer:      []byte{0x04, 0x04, 0x04, 0x00, 0x06, 0x00, 0x05, 0x8e, 0x87},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			port := &testSerialPort{
				readData: []byte(tt.fromServer),
			}
			client := newModbusClient(logger, port, 1*time.Minute)
			resp, err := client.ReadInputRegisters(0x04, 0, 2)
			if tt.fromServerError != nil {
				assert.Equal(t, tt.fromServerError, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.toServer, port.writeData)
			assert.Equal(t, tt.registers, resp)
		})
	}
}

func TestRTUWriteSingleCoil(t *testing.T) {
	tests := []struct {
		name            string
		toServer        []byte
		registers       []uint16
		fromServerError error
		fromServer      []byte
	}{
		{
			name:       "Valid",
			toServer:   []byte{0x04, 0x05, 0x00, 0x0A, 0xFF, 0x00, 0xAC, 0x6D},
			registers:  []uint16{0x0006, 0x0005},
			fromServer: []byte{0x04, 0x05, 0x00, 0x0A, 0xFF, 0x00, 0xAC, 0x6D},
		},
		{
			name:            "ServerError_IllegalDataAddress",
			toServer:        []byte{0x04, 0x05, 0x00, 0x0A, 0xFF, 0x00, 0xAC, 0x6D},
			fromServer:      []byte{0x04, 0x85, 0x02, 0xD3, 0x50},
			fromServerError: common.ErrIllegalDataAddress,
		},
		{
			name:            "InvalidRequest_InvalidChecksum",
			toServer:        []byte{0x04, 0x05, 0x00, 0x0A, 0xFF, 0x00, 0xAC, 0x6D},
			fromServerError: common.ErrInvalidChecksum,
			fromServer:      []byte{0x04, 0x05, 0x00, 0x0A, 0xFF, 0x00, 0xAD, 0xDD},
		},
		{
			name:            "InvalidRequest_ResponseValueMismatch",
			toServer:        []byte{0x04, 0x05, 0x00, 0x0A, 0xFF, 0x00, 0xAC, 0x6D},
			fromServerError: common.ErrResponseValueMismatch,
			fromServer:      []byte{0x04, 0x05, 0x00, 0x0A, 0x00, 0x00, 0xED, 0x9D},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			port := &testSerialPort{
				readData: []byte(tt.fromServer),
			}
			client := newModbusClient(logger, port, 1*time.Minute)
			err := client.WriteSingleCoil(0x04, 10, true)
			if tt.fromServerError != nil {
				assert.Equal(t, tt.fromServerError, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.toServer, port.writeData)
		})
	}
}

func TestRTUWriteSingleRegister(t *testing.T) {
	tests := []struct {
		name            string
		toServer        []byte
		register        uint16
		fromServerError error
		fromServer      []byte
	}{
		{
			name:       "Valid",
			toServer:   []byte{0x04, 0x06, 0x00, 0x10, 0x00, 0x03, 0xC8, 0x5B},
			register:   uint16(0x0003),
			fromServer: []byte{0x04, 0x06, 0x00, 0x10, 0x00, 0x03, 0xC8, 0x5B},
		},
		{
			name:            "ServerError_IllegalDataAddress",
			toServer:        []byte{0x04, 0x06, 0x00, 0x10, 0x00, 0x03, 0xC8, 0x5B},
			fromServer:      []byte{0x04, 0x86, 0x02, 0xD3, 0xA0},
			fromServerError: common.ErrIllegalDataAddress,
		},
		{
			name:            "InvalidRequest_InvalidChecksum",
			toServer:        []byte{0x04, 0x06, 0x00, 0x10, 0x00, 0x03, 0xC8, 0x5B},
			fromServerError: common.ErrInvalidChecksum,
			fromServer:      []byte{0x04, 0x06, 0x00, 0x10, 0x00, 0x03, 0xC8, 0x5C},
		},
		{
			name:            "InvalidRequest_ResponseValueMismatch",
			toServer:        []byte{0x04, 0x06, 0x00, 0x10, 0x00, 0x03, 0xC8, 0x5B},
			fromServerError: common.ErrResponseValueMismatch,
			fromServer:      []byte{0x04, 0x06, 0x00, 0x10, 0x00, 0x04, 0x89, 0x99},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			port := &testSerialPort{
				readData: []byte(tt.fromServer),
			}
			client := newModbusClient(logger, port, 1*time.Minute)
			err := client.WriteSingleRegister(0x04, 16, 3)
			if tt.fromServerError != nil {
				assert.Equal(t, tt.fromServerError, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.toServer, port.writeData)
		})
	}
}

func TestRTUWriteMultipleCoils(t *testing.T) {
	tests := []struct {
		name            string
		toServer        []byte
		coils           []bool
		fromServerError error
		fromServer      []byte
	}{
		{
			name:       "Valid",
			toServer:   []byte{0x04, 0x0F, 0x00, 0x00, 0x00, 0x18, 0x03, 0x01, 0x83, 0x07, 0x21, 0x56},
			coils:      []bool{true, false, false, false, false, false, false, false, true, true, false, false, false, false, false, true, true, true, true, false, false, false, false, false},
			fromServer: []byte{0x04, 0x0F, 0x00, 0x00, 0x00, 0x18, 0x55, 0x94},
		},
		{
			name:            "ServerError_IllegalDataAddress",
			toServer:        []byte{0x04, 0x0F, 0x00, 0x00, 0x00, 0x18, 0x03, 0x01, 0x83, 0x07, 0x21, 0x56},
			fromServer:      []byte{0x04, 0x8F, 0x02, 0xD5, 0xF0},
			fromServerError: common.ErrIllegalDataAddress,
		},
		{
			name:            "InvalidRequest_InvalidChecksum",
			toServer:        []byte{0x04, 0x0F, 0x00, 0x00, 0x00, 0x18, 0x03, 0x01, 0x83, 0x07, 0x21, 0x56},
			fromServerError: common.ErrInvalidChecksum,
			fromServer:      []byte{0x04, 0x0F, 0x00, 0x00, 0x00, 0x18, 0x55, 0x93},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			port := &testSerialPort{
				readData: []byte(tt.fromServer),
			}
			client := newModbusClient(logger, port, 1*time.Minute)
			err := client.WriteMultipleCoils(0x04, 0, tt.coils)
			if tt.fromServerError != nil {
				assert.Equal(t, tt.fromServerError, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.toServer, port.writeData)
		})
	}
}

func TestRTUWriteMultipleRegisters(t *testing.T) {
	tests := []struct {
		name            string
		toServer        []byte
		registers       []uint16
		fromServerError error
		fromServer      []byte
	}{
		{
			name:       "Valid",
			toServer:   []byte{0x04, 0x10, 0x00, 0x00, 0x00, 0x02, 0x04, 0x00, 0x04, 0x00, 0x02, 0x22, 0x63},
			registers:  []uint16{0x0004, 0x0002},
			fromServer: []byte{0x04, 0x10, 0x00, 0x00, 0x00, 0x02, 0x41, 0x9D},
		},
		{
			name:            "ServerError_IllegalDataAddress",
			toServer:        []byte{0x04, 0x10, 0x00, 0x00, 0x00, 0x02, 0x04, 0x00, 0x04, 0x00, 0x02, 0x22, 0x63},
			fromServer:      []byte{0x04, 0x90, 0x02, 0xDD, 0xC0},
			fromServerError: common.ErrIllegalDataAddress,
		},
		{
			name:            "InvalidRequest_InvalidChecksum",
			toServer:        []byte{0x04, 0x10, 0x00, 0x00, 0x00, 0x02, 0x04, 0x00, 0x04, 0x00, 0x02, 0x22, 0x63},
			fromServerError: common.ErrInvalidChecksum,
			fromServer:      []byte{0x04, 0x10, 0x00, 0x00, 0x00, 0x02, 0x41, 0x91},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			port := &testSerialPort{
				readData: []byte(tt.fromServer),
			}
			client := newModbusClient(logger, port, 1*time.Minute)
			err := client.WriteMultipleRegisters(0x04, 0, tt.registers)
			if tt.fromServerError != nil {
				assert.Equal(t, tt.fromServerError, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.toServer, port.writeData)
		})
	}
}
