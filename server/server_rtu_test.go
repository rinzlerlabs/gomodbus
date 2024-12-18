package server

import (
	"io"
	"testing"

	"github.com/rinzlerlabs/gomodbus/common"
	"github.com/rinzlerlabs/gomodbus/data"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

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

func TestRTUAcceptRequest(t *testing.T) {
	logger := zaptest.NewLogger(t)
	port := &testSerialPort{
		readData: []byte{0x04, 0x01, 0x00, 0x0A, 0x00, 0x0D, 0xDD, 0x98},
	}
	s, e := newModbusRTUServerWithHandler(logger, port, 0x04, NewDefaultHandler(logger, 1024, 1024, 1024, 1024))
	assert.NoError(t, e)
	adu, err := s.acceptAndValidateRequest()
	assert.NoError(t, err)
	assert.NotNil(t, adu)
	assert.Equal(t, uint16(0x04), adu.Address())
	assert.Equal(t, data.FunctionCode(0x01), adu.PDU().Function)
	assert.Equal(t, []byte{0x00, 0x0A, 0x00, 0x0D}, adu.PDU().Data)
	assert.Equal(t, []byte{0xDD, 0x98}, adu.Checksum())
}

func TestRTUReadCoils(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name         string
		request      []byte
		coils        []bool
		readError    error
		handlerError error
		response     []byte
	}{
		{
			name:     "Valid",
			request:  []byte{0x04, 0x01, 0x00, 0x0A, 0x00, 0x0D, 0xDD, 0x98},
			coils:    []bool{true, true, false, false, false, false, false, false, false, false, false, false, true, false, true, false, false, false, false, true, false, false, false, true, true, true, true},
			response: []byte{0x04, 0x01, 0x02, 0x0A, 0x11, 0xB3, 0x50},
		},
		{
			name:      "InvalidRequest_IvalidChecksum",
			request:   []byte{0x04, 0x01, 0x00, 0x0A, 0x00, 0x0D, 0xDD, 0x99},
			readError: common.ErrInvalidChecksum,
		},
		{
			name:      "InvalidRequest_NotOurAddress",
			request:   []byte{0x05, 0x01, 0x00, 0x0A, 0x00, 0x0D, 0xdc, 0x49},
			readError: common.ErrNotOurAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			s, e := newModbusRTUServerWithHandler(logger, port, 0x04, NewDefaultHandler(logger, 1024, 1024, 1024, 1024))
			assert.NoError(t, e)

			s.handler.(*DefaultHandler).Coils = tt.coils
			adu, err := s.acceptAndValidateRequest()
			if tt.readError != nil {
				assert.Error(t, err)
				assert.Nil(t, adu)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, adu)
			err = s.handler.Handle(adu)
			if tt.handlerError != nil {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.response, port.writeData)
			}
		})
	}
}

func TestRTUReadDiscreteInputs(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name         string
		request      []byte
		inputs       []bool
		readError    error
		handlerError error
		response     []byte
	}{
		{
			name:     "Valid",
			request:  []byte{0x04, 0x02, 0x00, 0x0A, 0x00, 0x0D, 0x99, 0x98},
			inputs:   []bool{true, true, false, false, false, false, false, false, false, false, false, false, true, false, true, false, false, false, false, true, false, false, false, true, true, true, true},
			response: []byte{0x04, 0x02, 0x02, 0x0A, 0x11, 0xb3, 0x14},
		},
		{
			name:      "InvalidRequest_NotOurAddress",
			request:   []byte{0x05, 0x02, 0x00, 0x0A, 0x00, 0x0D, 0x98, 0x49},
			readError: common.ErrNotOurAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			s, e := newModbusRTUServerWithHandler(logger, port, 0x04, NewDefaultHandler(logger, 1024, 1024, 1024, 1024))
			assert.NoError(t, e)

			s.handler.(*DefaultHandler).DiscreteInputs = tt.inputs
			adu, err := s.acceptAndValidateRequest()
			if tt.readError != nil {
				assert.Error(t, err)
				assert.Nil(t, adu)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, adu)
			err = s.handler.Handle(adu)
			if tt.handlerError != nil {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.response, port.writeData)
			}
		})
	}
}

func TestRTUReadHoldingRegisters(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name         string
		request      []byte
		registers    []uint16
		readError    error
		handlerError error
		response     []byte
	}{
		{
			name:      "Valid",
			request:   []byte{0x04, 0x03, 0x00, 0x00, 0x00, 0x02, 0xc4, 0x5e},
			registers: []uint16{0x0007, 0x0006, 0x0005, 0x0004, 0x0003, 0x0002, 0x0001, 0x0000},
			response:  []byte{0x04, 0x03, 0x04, 0x00, 0x06, 0x00, 0x05, 0x8f, 0x31},
		},
		{
			name:      "InvalidRequest_NotOurAddress",
			request:   []byte{0x05, 0x03, 0x00, 0x00, 0x00, 0x02, 0xc5, 0x8f},
			readError: common.ErrNotOurAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			s, e := newModbusRTUServerWithHandler(logger, port, 0x04, NewDefaultHandler(logger, 1024, 1024, 1024, 1024))
			assert.NoError(t, e)

			s.handler.(*DefaultHandler).HoldingRegisters = tt.registers
			adu, err := s.acceptAndValidateRequest()
			if tt.readError != nil {
				assert.Error(t, err)
				assert.Nil(t, adu)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, adu)
			err = s.handler.Handle(adu)
			if tt.handlerError != nil {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.response, port.writeData)
			}
		})
	}
}

func TestRTUReadInputRegisters(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name         string
		request      []byte
		registers    []uint16
		readError    error
		handlerError error
		response     []byte
	}{
		{
			name:      "Valid",
			request:   []byte{0x04, 0x04, 0x00, 0x00, 0x00, 0x02, 0x71, 0x9e},
			registers: []uint16{0x0007, 0x0006, 0x0005, 0x0004, 0x0003, 0x0002, 0x0001, 0x0000},
			response:  []byte{0x04, 0x04, 0x04, 0x00, 0x06, 0x00, 0x05, 0x8e, 0x86},
		},
		{
			name:      "InvalidRequest_NotOurAddress",
			request:   []byte{0x05, 0x04, 0x00, 0x00, 0x00, 0x02, 0x70, 0x4f},
			readError: common.ErrNotOurAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			s, e := newModbusRTUServerWithHandler(logger, port, 0x04, NewDefaultHandler(logger, 1024, 1024, 1024, 1024))
			assert.NoError(t, e)

			s.handler.(*DefaultHandler).InputRegisters = tt.registers
			adu, err := s.acceptAndValidateRequest()
			if tt.readError != nil {
				assert.Error(t, err)
				assert.Nil(t, adu)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, adu)
			err = s.handler.Handle(adu)
			if tt.handlerError != nil {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.response, port.writeData)
			}
		})
	}
}

func TestRTUWriteSingleCoil(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name         string
		request      []byte
		coilIndex    uint16
		readError    error
		handlerError error
		response     []byte
	}{
		{
			name:      "Valid",
			request:   []byte{0x04, 0x05, 0x00, 0x0A, 0xFF, 0x00, 0xAC, 0x6D},
			coilIndex: 10,
			response:  []byte{0x04, 0x05, 0x00, 0x0A, 0xFF, 0x00, 0xAC, 0x6D},
		},
		{
			name:      "InvalidRequest_NotOurAddress",
			request:   []byte{0x05, 0x05, 0x00, 0x0A, 0xFF, 0x00, 0xAD, 0xBC},
			readError: common.ErrNotOurAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			s, e := newModbusRTUServerWithHandler(logger, port, 0x04, NewDefaultHandler(logger, 1024, 1024, 1024, 1024))
			assert.NoError(t, e)

			adu, err := s.acceptAndValidateRequest()
			if tt.readError != nil {
				assert.Error(t, err)
				assert.Nil(t, adu)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, adu)
			err = s.handler.Handle(adu)
			if tt.handlerError != nil {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.response, port.writeData)
				assert.True(t, s.handler.(*DefaultHandler).Coils[tt.coilIndex+1])
			}
		})
	}
}

func TestRTUWriteSingleRegister(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name          string
		request       []byte
		registerIndex uint16
		readError     error
		handlerError  error
		response      []byte
	}{
		{
			name:          "Valid",
			request:       []byte{0x04, 0x06, 0x00, 0x10, 0x00, 0x03, 0xC8, 0x5B},
			registerIndex: 0x10,
			response:      []byte{0x04, 0x06, 0x00, 0x10, 0x00, 0x03, 0xC8, 0x5B},
		},
		{
			name:      "InvalidRequest_NotOurAddress",
			request:   []byte{0x05, 0x06, 0x00, 0x10, 0x00, 0x03, 0xC9, 0x8A},
			readError: common.ErrNotOurAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			s, e := newModbusRTUServerWithHandler(logger, port, 0x04, NewDefaultHandler(logger, 1024, 1024, 1024, 1024))
			assert.NoError(t, e)

			adu, err := s.acceptAndValidateRequest()
			if tt.readError != nil {
				assert.Error(t, err)
				assert.Nil(t, adu)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, adu)
			err = s.handler.Handle(adu)
			if tt.handlerError != nil {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.response, port.writeData)
				assert.Equal(t, uint16(3), s.handler.(*DefaultHandler).HoldingRegisters[tt.registerIndex+1])
			}
		})
	}
}

func TestRTUWriteMultipleCoils(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name              string
		request           []byte
		expectedRegisters []bool
		readError         error
		handlerError      error
		response          []byte
	}{
		{
			name:              "Valid",
			request:           []byte{0x04, 0x0F, 0x00, 0x00, 0x00, 0x18, 0x03, 0x01, 0x83, 0x07, 0x21, 0x56},
			expectedRegisters: []bool{true, false, false, false, false, false, false, false, true, true, false, false, false, false, false, true, true, true, true, false, false, false, false, false},
			response:          []byte{0x04, 0x0F, 0x00, 0x00, 0x00, 0x18, 0x55, 0x94},
		},
		{
			name:      "InvalidRequest_NotOurAddress",
			request:   []byte{0x05, 0x0F, 0x00, 0x00, 0x00, 0x18, 0x03, 0x01, 0x83, 0x07, 0x70, 0x93},
			readError: common.ErrNotOurAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			s, e := newModbusRTUServerWithHandler(logger, port, 0x04, NewDefaultHandler(logger, 1024, 1024, 1024, 1024))
			assert.NoError(t, e)

			adu, err := s.acceptAndValidateRequest()
			if tt.readError != nil {
				assert.Error(t, err)
				assert.Nil(t, adu)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, adu)
			err = s.handler.Handle(adu)
			if tt.handlerError != nil {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.response, port.writeData)
				assert.Equal(t, tt.expectedRegisters, s.handler.(*DefaultHandler).Coils[1:len(tt.expectedRegisters)+1])
			}
		})
	}
}

func TestRTUWriteMultipleRegisters(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name              string
		request           []byte
		expectedRegisters []uint16
		readError         error
		handlerError      error
		response          []byte
	}{
		{
			name:              "Valid",
			request:           []byte{0x04, 0x10, 0x00, 0x00, 0x00, 0x02, 0x04, 0x00, 0x04, 0x00, 0x02, 0x22, 0x63},
			expectedRegisters: []uint16{0x0004, 0x0002, 0x0000, 0x0000},
			response:          []byte{0x04, 0x10, 0x00, 0x00, 0x00, 0x02, 0x41, 0x9D},
		},
		{
			name:      "InvalidRequest_NotOurAddress",
			request:   []byte{0x05, 0x10, 0x00, 0x00, 0x00, 0x02, 0x04, 0x00, 0x04, 0x00, 0x02, 0x26, 0x9F},
			readError: common.ErrNotOurAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			s, e := newModbusRTUServerWithHandler(logger, port, 0x04, NewDefaultHandler(logger, 1024, 1024, 1024, 1024))
			assert.NoError(t, e)

			adu, err := s.acceptAndValidateRequest()
			if tt.readError != nil {
				assert.Error(t, err)
				assert.Nil(t, adu)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, adu)
			err = s.handler.Handle(adu)
			if tt.handlerError != nil {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.response, port.writeData)
				assert.Equal(t, tt.expectedRegisters, s.handler.(*DefaultHandler).HoldingRegisters[1:len(tt.expectedRegisters)+1])
			}
		})
	}
}

func TestCreateCrc(t *testing.T) {
	bytes := []byte{0x04, 0x06, 0x00, 0x10, 0x00, 0x04}
	t.Logf("CRC: %x", calculateCrc(bytes))
}

func calculateCrc(data []byte) uint16 {
	var crc uint16 = 0xFFFF
	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if (crc & 1) != 0 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}
