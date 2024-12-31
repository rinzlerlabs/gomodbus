package rtu

import (
	"io"
	"sync"
	"testing"
	"time"

	"github.com/rinzlerlabs/gomodbus/server"
	"github.com/rinzlerlabs/gomodbus/server/serial"
	"github.com/rinzlerlabs/gomodbus/transport/serial/rtu"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func newModbusServerWithHandler(logger *zap.Logger, stream io.ReadWriteCloser, serverAddress uint16, handler server.RequestHandler) (serial.ModbusSerialServer, error) {
	return serial.NewModbusSerialServerWithTransport(logger, serverAddress, handler, rtu.NewModbusServerTransport(stream, logger, serverAddress))
}

type testSerialPort struct {
	firstReadDone bool
	mu            sync.Mutex
	readData      []byte
	writeData     []byte
}

func (t *testSerialPort) Read(b []byte) (n int, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.firstReadDone {
		time.Sleep(100 * time.Millisecond)
		copy(b, []byte{0x00})
		t.firstReadDone = true
		return 1, nil
	}
	if len(t.readData) == 0 {
		return 0, nil
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

func waitForWrite(port *testSerialPort, desiredLength int) {
	timeout := time.After(100 * time.Second)
	tick := time.Tick(10 * time.Millisecond)

	for {
		select {
		case <-timeout:
			return
		case <-tick:
			if len(port.writeData) == desiredLength {
				return
			}
		}
	}
}

func TestNilHandlerReturnsError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	port := &testSerialPort{}
	_, err := newModbusServerWithHandler(logger, port, 0x04, nil)
	assert.Error(t, err)
}

func TestAcceptRequest(t *testing.T) {
	logger := zaptest.NewLogger(t)
	port := &testSerialPort{
		readData: []byte{0x04, 0x01, 0x00, 0x0A, 0x00, 0x0D, 0xDD, 0x98},
	}
	s, err := newModbusServerWithHandler(logger, port, 0x04, server.NewDefaultHandler(logger, 1024, 1024, 1024, 1024))
	assert.NoError(t, err)

	s.Start()
	assert.NoError(t, err)

	waitForWrite(port, 7)

	err = s.Close()
	assert.NoError(t, err)

	adu := port.writeData

	assert.Equal(t, []byte{0x04, 0x01, 0x02, 0x00, 0x00, 0x75, 0xFC}, adu)
}

func TestReadCoils(t *testing.T) {
	tests := []struct {
		name     string
		request  []byte
		response []byte
		coils    []bool
	}{
		{
			name:     "Valid",
			request:  []byte{0x04, 0x01, 0x00, 0x0A, 0x00, 0x0D, 0xDD, 0x98},
			response: []byte{0x04, 0x01, 0x02, 0x0A, 0x11, 0xB3, 0x50},
			coils:    []bool{true, false, false, false, false, false, false, false, false, false, false, true, false, true, false, false, false, false, true, false, false, false, true, true, true, true},
		},
		{
			name:    "InvalidRequest_IvalidChecksum",
			request: []byte{0x04, 0x01, 0x00, 0x0A, 0x00, 0x0D, 0xDD, 0x99},
		},
		{
			name:    "InvalidRequest_NotOurAddress",
			request: []byte{0x05, 0x01, 0x00, 0x0A, 0x00, 0x0D, 0xDC, 0x49},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			handler := server.NewDefaultHandler(logger, 1024, 1024, 1024, 1024)
			if tt.coils != nil {
				handler.(*server.DefaultHandler).Coils = tt.coils
			}
			s, err := newModbusServerWithHandler(logger, port, 0x04, handler)
			assert.NoError(t, err)

			s.Start()
			assert.NoError(t, err)

			waitForWrite(port, len(tt.response))

			err = s.Close()
			assert.NoError(t, err)
			assert.Equal(t, tt.response, port.writeData)
		})
	}
}

func TestReadDiscreteInputs(t *testing.T) {
	tests := []struct {
		name     string
		request  []byte
		inputs   []bool
		response []byte
	}{
		{
			name:     "Valid",
			request:  []byte{0x04, 0x02, 0x00, 0x0A, 0x00, 0x0D, 0x99, 0x98},
			response: []byte{0x04, 0x02, 0x02, 0x0A, 0x11, 0xb3, 0x14},
			inputs:   []bool{true, false, false, false, false, false, false, false, false, false, false, true, false, true, false, false, false, false, true, false, false, false, true, true, true, true},
		},
		{
			name:    "InvalidRequest_IvalidChecksum",
			request: []byte{0x04, 0x02, 0x00, 0x0A, 0x00, 0x0D, 0x99, 0x99},
		},
		{
			name:    "InvalidRequest_NotOurAddress",
			request: []byte{0x05, 0x02, 0x00, 0x0A, 0x00, 0x0D, 0x98, 0x49},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			handler := server.NewDefaultHandler(logger, 1024, 1024, 1024, 1024)
			if tt.inputs != nil {
				handler.(*server.DefaultHandler).DiscreteInputs = tt.inputs
			}
			s, err := newModbusServerWithHandler(logger, port, 0x04, handler)
			assert.NoError(t, err)

			s.Start()
			assert.NoError(t, err)

			waitForWrite(port, len(tt.response))

			err = s.Close()
			assert.NoError(t, err)
			assert.Equal(t, tt.response, port.writeData)
		})
	}
}

func TestReadHoldingRegisters(t *testing.T) {
	tests := []struct {
		name      string
		request   []byte
		response  []byte
		registers []uint16
	}{
		{
			name:      "Valid",
			request:   []byte{0x04, 0x03, 0x00, 0x00, 0x00, 0x02, 0xC4, 0x5E},
			response:  []byte{0x04, 0x03, 0x04, 0x00, 0x06, 0x00, 0x05, 0x8f, 0x31},
			registers: []uint16{0x0006, 0x0005, 0x0004, 0x0003, 0x0002, 0x0001, 0x0000},
		},
		{
			name:    "InvalidRequest_IvalidChecksum",
			request: []byte{0x04, 0x03, 0x00, 0x00, 0x00, 0x02, 0xC4, 0x5F},
		},
		{
			name:    "InvalidRequest_NotOurAddress",
			request: []byte{0x05, 0x03, 0x00, 0x00, 0x00, 0x02, 0xC5, 0x8F},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			handler := server.NewDefaultHandler(logger, 1024, 1024, 1024, 1024)
			if tt.registers != nil {
				handler.(*server.DefaultHandler).HoldingRegisters = tt.registers
			}
			s, err := newModbusServerWithHandler(logger, port, 0x04, handler)
			assert.NoError(t, err)

			s.Start()
			assert.NoError(t, err)

			waitForWrite(port, len(tt.response))

			err = s.Close()
			assert.NoError(t, err)
			assert.Equal(t, tt.response, port.writeData)
		})
	}
}

func TestReadInputRegisters(t *testing.T) {
	tests := []struct {
		name      string
		request   []byte
		response  []byte
		registers []uint16
	}{
		{
			name:      "Valid",
			request:   []byte{0x04, 0x04, 0x00, 0x00, 0x00, 0x02, 0x71, 0x9E},
			registers: []uint16{0x0006, 0x0005, 0x0004, 0x0003, 0x0002, 0x0001, 0x0000},
			response:  []byte{0x04, 0x04, 0x04, 0x00, 0x06, 0x00, 0x05, 0x8E, 0x86},
		},
		{
			name:    "InvalidRequest_IvalidChecksum",
			request: []byte{0x04, 0x04, 0x00, 0x00, 0x00, 0x02, 0x71, 0x9F},
		},
		{
			name:    "InvalidRequest_NotOurAddress",
			request: []byte{0x05, 0x04, 0x00, 0x00, 0x00, 0x02, 0x70, 0x4F},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			handler := server.NewDefaultHandler(logger, 1024, 1024, 1024, 1024)
			if tt.registers != nil {
				handler.(*server.DefaultHandler).InputRegisters = tt.registers
			}
			s, err := newModbusServerWithHandler(logger, port, 0x04, handler)
			assert.NoError(t, err)

			s.Start()
			assert.NoError(t, err)

			waitForWrite(port, len(tt.response))

			err = s.Close()
			assert.NoError(t, err)
			assert.Equal(t, tt.response, port.writeData)
		})
	}
}

func TestWriteSingleCoil(t *testing.T) {
	tests := []struct {
		name      string
		request   []byte
		response  []byte
		coilIndex uint16
		coilValue bool
	}{
		{
			name:      "Valid",
			request:   []byte{0x04, 0x05, 0x00, 0x0A, 0xFF, 0x00, 0xAC, 0x6D},
			response:  []byte{0x04, 0x05, 0x00, 0x0A, 0xFF, 0x00, 0xAC, 0x6D},
			coilIndex: 10,
			coilValue: true,
		},

		{
			name:      "InvalidRequest_InvalidChecksum",
			request:   []byte{0x04, 0x05, 0x00, 0x0A, 0xFF, 0x00, 0xAC, 0x6F},
			coilIndex: 10,
		},
		{
			name:    "InvalidRequest_NotOurAddress",
			request: []byte{0x05, 0x05, 0x00, 0x0A, 0xFF, 0x00, 0xAD, 0xBC},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			handler := server.NewDefaultHandler(logger, 1024, 1024, 1024, 1024)
			s, err := newModbusServerWithHandler(logger, port, 0x04, handler)
			assert.NoError(t, err)

			s.Start()
			assert.NoError(t, err)

			waitForWrite(port, len(tt.response))

			err = s.Close()
			assert.NoError(t, err)
			if tt.coilIndex > 0 {
				assert.Equal(t, tt.coilValue, handler.(*server.DefaultHandler).Coils[tt.coilIndex+1])
			}
			assert.Equal(t, tt.response, port.writeData)
		})
	}
}

func TestWriteSingleRegister(t *testing.T) {
	tests := []struct {
		name          string
		request       []byte
		response      []byte
		registerIndex uint16
		registerValue uint16
	}{
		{
			name:          "Valid",
			request:       []byte{0x04, 0x06, 0x00, 0x10, 0x00, 0x03, 0xC8, 0x5B},
			response:      []byte{0x04, 0x06, 0x00, 0x10, 0x00, 0x03, 0xC8, 0x5B},
			registerIndex: 0x10,
			registerValue: 0x0003,
		},
		{
			name:          "InvalidRequest_InvalidChecksum",
			request:       []byte{0x04, 0x06, 0x00, 0x10, 0x00, 0x03, 0xC8, 0x5C},
			registerIndex: 0x10,
		},
		{
			name:    "InvalidRequest_NotOurAddress",
			request: []byte{0x05, 0x06, 0x00, 0x10, 0x00, 0x03, 0xC9, 0x8A},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			handler := server.NewDefaultHandler(logger, 1024, 1024, 1024, 1024)
			s, err := newModbusServerWithHandler(logger, port, 0x04, handler)
			assert.NoError(t, err)

			s.Start()
			assert.NoError(t, err)

			waitForWrite(port, len(tt.response))

			err = s.Close()
			assert.NoError(t, err)
			if tt.registerIndex > 0 {
				assert.Equal(t, tt.registerValue, handler.(*server.DefaultHandler).HoldingRegisters[tt.registerIndex+1])
			}
			assert.Equal(t, tt.response, port.writeData)
		})
	}
}

func TestWriteMultipleCoils(t *testing.T) {
	tests := []struct {
		name              string
		request           []byte
		response          []byte
		expectedRegisters []bool
	}{
		{
			name:              "Valid",
			request:           []byte{0x04, 0x0F, 0x00, 0x00, 0x00, 0x18, 0x03, 0x01, 0x83, 0x07, 0x21, 0x56},
			response:          []byte{0x04, 0x0F, 0x00, 0x00, 0x00, 0x18, 0x55, 0x94},
			expectedRegisters: []bool{true, false, false, false, false, false, false, false, true, true, false, false, false, false, false, true, true, true, true, false, false, false, false, false},
		},
		{
			name:    "InvalidRequest_InvalidChecksum",
			request: []byte{0x04, 0x0F, 0x00, 0x00, 0x00, 0x18, 0x03, 0x01, 0x83, 0x07, 0x21, 0x56},
		},
		{
			name:    "InvalidRequest_NotOurAddress",
			request: []byte{0x05, 0x0F, 0x00, 0x00, 0x00, 0x18, 0x03, 0x01, 0x83, 0x07, 0x70, 0x93},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			handler := server.NewDefaultHandler(logger, 1024, 1024, 1024, 1024)
			s, err := newModbusServerWithHandler(logger, port, 0x04, handler)
			assert.NoError(t, err)

			s.Start()
			assert.NoError(t, err)

			waitForWrite(port, len(tt.response))

			err = s.Close()
			assert.NoError(t, err)
			if tt.expectedRegisters != nil {
				assert.Equal(t, tt.expectedRegisters, handler.(*server.DefaultHandler).Coils[0:24])
			}
			assert.Equal(t, tt.response, port.writeData)
		})
	}
}

func TestWriteMultipleRegisters(t *testing.T) {
	tests := []struct {
		name              string
		request           []byte
		response          []byte
		expectedRegisters []uint16
	}{
		{
			name:              "Valid",
			request:           []byte{0x04, 0x10, 0x00, 0x00, 0x00, 0x02, 0x04, 0x00, 0x04, 0x00, 0x02, 0x22, 0x63},
			response:          []byte{0x04, 0x10, 0x00, 0x00, 0x00, 0x02, 0x41, 0x9D},
			expectedRegisters: []uint16{0x0004, 0x0002},
		},
		{
			name:    "InvalidRequest_InvalidChecksum",
			request: []byte{0x04, 0x10, 0x00, 0x00, 0x00, 0x02, 0x04, 0x00, 0x04, 0x00, 0x02, 0x22, 0x64},
		},
		{
			name:    "InvalidRequest_NotOurAddress",
			request: []byte{0x05, 0x10, 0x00, 0x00, 0x00, 0x02, 0x04, 0x00, 0x04, 0x00, 0x02, 0x26, 0x9F},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			handler := server.NewDefaultHandler(logger, 1024, 1024, 1024, 1024)
			s, err := newModbusServerWithHandler(logger, port, 0x04, handler)
			assert.NoError(t, err)

			s.Start()
			assert.NoError(t, err)

			waitForWrite(port, len(tt.response))

			err = s.Close()
			assert.NoError(t, err)
			if tt.expectedRegisters != nil {
				assert.Equal(t, tt.expectedRegisters, handler.(*server.DefaultHandler).HoldingRegisters[0:2])
			}
			assert.Equal(t, tt.response, port.writeData)
		})
	}
}
