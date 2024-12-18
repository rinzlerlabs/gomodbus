package client

import (
	"testing"
	"time"

	"github.com/rinzlerlabs/gomodbus/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestASCIIReadCoils(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name            string
		toServer        string
		coils           []bool
		fromServerError error
		fromServer      string
	}{
		{
			name:       "Valid",
			toServer:   ":0401000A000DE4\r\n",
			coils:      []bool{false, true, false, true, false, false, false, false, true, false, false, false, true},
			fromServer: ":0401020A11DE\r\n",
		},
		{
			name:            "InvalidRequest_IvalidChecksum",
			toServer:        ":0401000A000DE4\r\n",
			fromServerError: common.ErrInvalidChecksum,
			fromServer:      ":0401000A000DE5\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := &testSerialPort{
				readData: []byte(tt.fromServer),
			}
			client := newModbusASCIIClient(logger, port, 1*time.Minute)
			resp, err := client.ReadCoils(0x04, 10, 13)
			if tt.fromServerError != nil {
				assert.Equal(t, tt.fromServerError, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.toServer, string(port.writeData))
			assert.Equal(t, tt.coils, resp)
		})
	}
}

func TestASCIIReadDiscreteInputs(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name            string
		toServer        string
		inputs          []bool
		fromServerError error
		fromServer      string
	}{
		{
			name:       "Valid",
			toServer:   ":0402000A000DE3\r\n",
			inputs:     []bool{false, true, false, true, false, false, false, false, true, false, false, false, true},
			fromServer: ":0402020A11DD\r\n",
		},
		{
			name:            "InvalidRequest_InvalidChecksum",
			toServer:        ":0402000A000DE3\r\n",
			fromServerError: common.ErrInvalidChecksum,
			fromServer:      ":0402020A11DA\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := &testSerialPort{
				readData: []byte(tt.fromServer),
			}
			client := newModbusASCIIClient(logger, port, 1*time.Minute)
			resp, err := client.ReadDiscreteInputs(0x04, 10, 13)
			if tt.fromServerError != nil {
				assert.Equal(t, tt.fromServerError, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.toServer, string(port.writeData))
			assert.Equal(t, tt.inputs, resp)
		})
	}
}

func TestASCIIReadHoldingRegisters(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name            string
		toServer        string
		registers       []uint16
		fromServerError error
		fromServer      string
	}{
		{
			name:       "Valid",
			toServer:   ":040300000002F7\r\n",
			registers:  []uint16{0x0006, 0x0005},
			fromServer: ":04030400060005EA\r\n",
		},
		{
			name:            "InvalidRequest_InvalidChecksum",
			toServer:        ":040300000002F7\r\n",
			fromServerError: common.ErrInvalidChecksum,
			fromServer:      ":04030400060005EB\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := &testSerialPort{
				readData: []byte(tt.fromServer),
			}
			client := newModbusASCIIClient(logger, port, 1*time.Minute)
			resp, err := client.ReadHoldingRegisters(0x04, 0, 2)
			if tt.fromServerError != nil {
				assert.Equal(t, tt.fromServerError, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.toServer, string(port.writeData))
			assert.Equal(t, tt.registers, resp)
		})
	}
}

func TestASCIIReadInputRegisters(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name            string
		toServer        string
		registers       []uint16
		fromServerError error
		fromServer      string
	}{
		{
			name:       "Valid",
			toServer:   ":040400000002F6\r\n",
			registers:  []uint16{0x0006, 0x0005},
			fromServer: ":04040400060005E9\r\n",
		},
		{
			name:            "InvalidRequest_InvalidChecksum",
			toServer:        ":040400000002F6\r\n",
			fromServerError: common.ErrInvalidChecksum,
			fromServer:      ":04040400060005E8\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := &testSerialPort{
				readData: []byte(tt.fromServer),
			}
			client := newModbusASCIIClient(logger, port, 1*time.Minute)
			resp, err := client.ReadInputRegisters(0x04, 0, 2)
			if tt.fromServerError != nil {
				assert.Equal(t, tt.fromServerError, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.toServer, string(port.writeData))
			assert.Equal(t, tt.registers, resp)
		})
	}
}

func TestASCIIWriteSingleCoil(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name            string
		toServer        string
		registers       []uint16
		fromServerError error
		fromServer      string
	}{
		{
			name:       "Valid",
			toServer:   ":0405000AFF00EE\r\n",
			registers:  []uint16{0x0006, 0x0005},
			fromServer: ":0405000AFF00EE\r\n",
		},
		{
			name:            "InvalidRequest_InvalidChecksum",
			toServer:        ":0405000AFF00EE\r\n",
			fromServerError: common.ErrInvalidChecksum,
			fromServer:      ":0405000AFF00ED\r\n",
		},
		{
			name:            "InvalidRequest_ResponseValueMismatch",
			toServer:        ":0405000AFF00EE\r\n",
			fromServerError: common.ErrResponseValueMismatch,
			fromServer:      ":0405000A0000ED\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := &testSerialPort{
				readData: []byte(tt.fromServer),
			}
			client := newModbusASCIIClient(logger, port, 1*time.Minute)
			err := client.WriteSingleCoil(0x04, 10, true)
			if tt.fromServerError != nil {
				assert.Equal(t, tt.fromServerError, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.toServer, string(port.writeData))
		})
	}
}

func TestASCIIWriteSingleRegister(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name            string
		toServer        string
		register        uint16
		fromServerError error
		fromServer      string
	}{
		{
			name:       "Valid",
			toServer:   ":040600100003E3\r\n",
			register:   uint16(0x0003),
			fromServer: ":040600100003E3\r\n",
		},
		{
			name:            "InvalidRequest_InvalidChecksum",
			toServer:        ":040600100003E3\r\n",
			fromServerError: common.ErrInvalidChecksum,
			fromServer:      ":040600100003E2\r\n",
		},
		{
			name:            "InvalidRequest_ResponseValueMismatch",
			toServer:        ":040600100003E3\r\n",
			fromServerError: common.ErrResponseValueMismatch,
			fromServer:      ":040600100004E2\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := &testSerialPort{
				readData: []byte(tt.fromServer),
			}
			client := newModbusASCIIClient(logger, port, 1*time.Minute)
			err := client.WriteSingleRegister(0x04, 16, 3)
			if tt.fromServerError != nil {
				assert.Equal(t, tt.fromServerError, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.toServer, string(port.writeData))
		})
	}
}

func TestASCIIWriteMultipleCoils(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name            string
		toServer        string
		coils           []bool
		fromServerError error
		fromServer      string
	}{
		{
			name:       "Valid",
			toServer:   ":040F000000180301830747\r\n",
			coils:      []bool{true, false, false, false, false, false, false, false, true, true, false, false, false, false, false, true, true, true, true, false, false, false, false, false},
			fromServer: ":040F00000018D5\r\n",
		},
		{
			name:            "InvalidRequest_InvalidChecksum",
			toServer:        ":040F000000180301830747\r\n",
			fromServerError: common.ErrInvalidChecksum,
			fromServer:      ":040F00000018D4\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := &testSerialPort{
				readData: []byte(tt.fromServer),
			}
			client := newModbusASCIIClient(logger, port, 1*time.Minute)
			err := client.WriteMultipleCoils(0x04, 0, tt.coils)
			if tt.fromServerError != nil {
				assert.Equal(t, tt.fromServerError, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.toServer, string(port.writeData))
		})
	}
}

func TestASCIIWriteMultipleRegisters(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name            string
		toServer        string
		registers       []uint16
		fromServerError error
		fromServer      string
	}{
		{
			name:       "Valid",
			toServer:   ":0410000000020400040002E0\r\n",
			registers:  []uint16{0x0004, 0x0002},
			fromServer: ":041000000002EA\r\n",
		},
		{
			name:            "InvalidRequest_InvalidChecksum",
			toServer:        ":0410000000020400040002E0\r\n",
			fromServerError: common.ErrInvalidChecksum,
			fromServer:      ":041000000002EB\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := &testSerialPort{
				readData: []byte(tt.fromServer),
			}
			client := newModbusASCIIClient(logger, port, 1*time.Minute)
			err := client.WriteMultipleRegisters(0x04, 0, tt.registers)
			if tt.fromServerError != nil {
				assert.Equal(t, tt.fromServerError, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.toServer, string(port.writeData))
		})
	}
}
