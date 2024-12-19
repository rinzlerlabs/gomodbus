package server

import (
	"encoding/hex"
	"io"
	"testing"

	"github.com/rinzlerlabs/gomodbus/common"
	"github.com/rinzlerlabs/gomodbus/data"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestASCIIAcceptRequest(t *testing.T) {
	logger := zaptest.NewLogger(t)
	port := &testSerialPort{
		readData: []byte{0x3A, 0x30, 0x32, 0x30, 0x31, 0x30, 0x30, 0x32, 0x30, 0x30, 0x30, 0x30, 0x43, 0x44, 0x31, 0x0D, 0x0A},
	}
	s, e := newModbusASCIIServerWithHandler(logger, port, 0x02, NewDefaultHandler(logger, 1024, 1024, 1024, 1024))
	assert.NoError(t, e)
	txn, err := s.acceptAndValidateTransaction()
	assert.NoError(t, err)
	assert.NotNil(t, txn)
	adu := txn.Frame()
	assert.Equal(t, uint16(0x02), adu.Address())
	assert.Equal(t, data.FunctionCode(0x01), adu.PDU().FunctionCode())
	assert.Equal(t, []byte{0x00, 0x20, 0x00, 0xC}, adu.PDU().Operation().Bytes())
	assert.Equal(t, []byte{0xD1}, []byte(adu.Checksum()))
}

func TestASCIIReadCoils(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name         string
		request      string
		coils        []bool
		readError    error
		handlerError error
		response     string
	}{
		{
			name:     "Valid",
			request:  ":0401000A000DE4\r\n",
			coils:    []bool{true, true, false, false, false, false, false, false, false, false, false, false, true, false, true, false, false, false, false, true, false, false, false, true, true, true, true},
			response: ":0401020A11DE\r\n",
		},
		{
			name:      "InvalidRequest_MissingTrailers",
			request:   ":0401000A000DE4",
			readError: io.EOF,
		},
		{
			name:      "InvalidRequest_InvalidStart",
			request:   "0401000A000DE4\r\n",
			readError: common.ErrInvalidPacket,
		},
		{
			name:      "InvalidRequest_IvalidChecksum",
			request:   ":0401000A000DE5\r\n",
			readError: common.ErrInvalidChecksum,
		},
		{
			name:      "InvalidRequest_NotOurAddress",
			request:   ":0501000A000DE3\r\n",
			readError: common.ErrNotOurAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			s, e := newModbusASCIIServerWithHandler(logger, port, 0x04, NewDefaultHandler(logger, 1024, 1024, 1024, 1024))
			assert.NoError(t, e)

			s.handler.(*DefaultHandler).Coils = tt.coils
			adu, err := s.acceptAndValidateTransaction()
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
				str := string(port.writeData)
				assert.Equal(t, tt.response, str)
			}
		})
	}
}

func TestASCIIReadDiscreteInputs(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name         string
		request      string
		inputs       []bool
		readError    error
		handlerError error
		response     string
	}{
		{
			name:     "Valid",
			request:  ":0402000A000DE3\r\n",
			inputs:   []bool{true, true, false, false, false, false, false, false, false, false, false, false, true, false, true, false, false, false, false, true, false, false, false, true, true, true, true},
			response: ":0402020A11DD\r\n",
		},
		{
			name:      "InvalidRequest_MissingTrailers",
			request:   ":0402000A000DE3",
			readError: io.EOF,
		},
		{
			name:      "InvalidRequest_InvalidStart",
			request:   "0402000A000DE3\r\n",
			readError: common.ErrInvalidPacket,
		},
		{
			name:      "InvalidRequest_NotOurAddress",
			request:   ":0502000A000DE2\r\n",
			readError: common.ErrNotOurAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			s, e := newModbusASCIIServerWithHandler(logger, port, 0x04, NewDefaultHandler(logger, 1024, 1024, 1024, 1024))
			assert.NoError(t, e)

			s.handler.(*DefaultHandler).DiscreteInputs = tt.inputs
			adu, err := s.acceptAndValidateTransaction()
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
				str := string(port.writeData)
				assert.Equal(t, tt.response, str)
			}
		})
	}
}

func TestASCIIReadHoldingRegisters(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name         string
		request      string
		registers    []uint16
		readError    error
		handlerError error
		response     string
	}{
		{
			name:      "Valid",
			request:   ":040300000002F7\r\n",
			registers: []uint16{0x0007, 0x0006, 0x0005, 0x0004, 0x0003, 0x0002, 0x0001, 0x0000},
			response:  ":04030400060005EA\r\n",
		},
		{
			name:      "InvalidRequest_MissingTrailers",
			request:   ":040300000002F7",
			readError: io.EOF,
		},
		{
			name:      "InvalidRequest_InvalidStart",
			request:   "040300000002F7\r\n",
			readError: common.ErrInvalidPacket,
		},
		{
			name:      "InvalidRequest_NotOurAddress",
			request:   ":010300000002FA\r\n",
			readError: common.ErrNotOurAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			s, e := newModbusASCIIServerWithHandler(logger, port, 0x04, NewDefaultHandler(logger, 1024, 1024, 1024, 1024))
			assert.NoError(t, e)

			s.handler.(*DefaultHandler).HoldingRegisters = tt.registers
			adu, err := s.acceptAndValidateTransaction()
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
				str := string(port.writeData)
				assert.Equal(t, tt.response, str)
			}
		})
	}
}

func TestASCIIReadInputRegisters(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name         string
		request      string
		registers    []uint16
		readError    error
		handlerError error
		response     string
	}{
		{
			name:      "Valid",
			request:   ":040400000002F6\r\n",
			registers: []uint16{0x0007, 0x0006, 0x0005, 0x0004, 0x0003, 0x0002, 0x0001, 0x0000},
			response:  ":04040400060005E9\r\n",
		},
		{
			name:      "InvalidRequest_MissingTrailers",
			request:   ":040400000002F6",
			readError: io.EOF,
		},
		{
			name:      "InvalidRequest_InvalidStart",
			request:   "040400000002F6\r\n",
			readError: common.ErrInvalidPacket,
		},
		{
			name:      "InvalidRequest_NotOurAddress",
			request:   ":050400000002F5\r\n",
			readError: common.ErrNotOurAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			s, e := newModbusASCIIServerWithHandler(logger, port, 0x04, NewDefaultHandler(logger, 1024, 1024, 1024, 1024))
			assert.NoError(t, e)

			s.handler.(*DefaultHandler).InputRegisters = tt.registers
			adu, err := s.acceptAndValidateTransaction()
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
				str := string(port.writeData)
				assert.Equal(t, tt.response, str)
			}
		})
	}
}

func TestASCIIWriteSingleCoil(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name         string
		request      string
		coilIndex    uint16
		readError    error
		handlerError error
		response     string
	}{
		{
			name:      "Valid",
			request:   ":0405000AFF00EE\r\n",
			coilIndex: 10,
			response:  ":0405000AFF00EE\r\n",
		},
		{
			name:      "InvalidRequest_MissingTrailers",
			request:   ":0405000AFF00EE",
			readError: io.EOF,
		},
		{
			name:      "InvalidRequest_InvalidStart",
			request:   "0405000AFF00EE\r\n",
			readError: common.ErrInvalidPacket,
		},
		{
			name:      "InvalidRequest_NotOurAddress",
			request:   ":0505000AFF00ED\r\n",
			readError: common.ErrNotOurAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			s, e := newModbusASCIIServerWithHandler(logger, port, 0x04, NewDefaultHandler(logger, 1024, 1024, 1024, 1024))
			assert.NoError(t, e)

			adu, err := s.acceptAndValidateTransaction()
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
				str := string(port.writeData)
				assert.Equal(t, tt.response, str)
				assert.True(t, s.handler.(*DefaultHandler).Coils[tt.coilIndex+1])
			}
		})
	}
}

func TestASCIIWriteSingleRegister(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name          string
		request       string
		registerIndex uint16
		readError     error
		handlerError  error
		response      string
	}{
		{
			name:          "Valid",
			request:       ":040600100003E3\r\n",
			registerIndex: 0x10,
			response:      ":040600100003E3\r\n",
		},
		{
			name:      "InvalidRequest_MissingTrailers",
			request:   ":040600100003E3",
			readError: io.EOF,
		},
		{
			name:      "InvalidRequest_InvalidStart",
			request:   "040600100003E3\r\n",
			readError: common.ErrInvalidPacket,
		},
		{
			name:      "InvalidRequest_NotOurAddress",
			request:   ":050600100003E2\r\n",
			readError: common.ErrNotOurAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			s, e := newModbusASCIIServerWithHandler(logger, port, 0x04, NewDefaultHandler(logger, 1024, 1024, 1024, 1024))
			assert.NoError(t, e)

			adu, err := s.acceptAndValidateTransaction()
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
				str := string(port.writeData)
				assert.Equal(t, tt.response, str)
				assert.Equal(t, uint16(3), s.handler.(*DefaultHandler).HoldingRegisters[tt.registerIndex+1])
			}
		})
	}
}

func TestASCIIWriteMultipleCoils(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name              string
		request           string
		expectedRegisters []bool
		readError         error
		handlerError      error
		response          string
	}{
		{
			name:              "Valid",
			request:           ":040F000000180301830747\r\n",
			expectedRegisters: []bool{true, false, false, false, false, false, false, false, true, true, false, false, false, false, false, true, true, true, true, false, false, false, false, false},
			response:          ":040F00000018D5\r\n",
		},
		{
			name:      "InvalidRequest_MissingTrailers",
			request:   ":040F000000180301830747",
			readError: io.EOF,
		},
		{
			name:      "InvalidRequest_InvalidStart",
			request:   "040F000000180301830747\r\n",
			readError: common.ErrInvalidPacket,
		},
		{
			name:      "InvalidRequest_NotOurAddress",
			request:   ":050F000000180301830746\r\n",
			readError: common.ErrNotOurAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			s, e := newModbusASCIIServerWithHandler(logger, port, 0x04, NewDefaultHandler(logger, 1024, 1024, 1024, 1024))
			assert.NoError(t, e)

			adu, err := s.acceptAndValidateTransaction()
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
				str := string(port.writeData)
				assert.Equal(t, tt.response, str)
				assert.Equal(t, tt.expectedRegisters, s.handler.(*DefaultHandler).Coils[1:len(tt.expectedRegisters)+1])
			}
		})
	}
}

func TestASCIIWriteMultipleRegisters(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name              string
		request           string
		expectedRegisters []uint16
		readError         error
		handlerError      error
		response          string
	}{
		{
			name:              "Valid",
			request:           ":0410000000020400040002E0\r\n",
			expectedRegisters: []uint16{0x0004, 0x0002, 0x0000, 0x0000},
			response:          ":041000000002EA\r\n",
		},
		{
			name:      "InvalidRequest_MissingTrailers",
			request:   ":0410000000020400040002E0",
			readError: io.EOF,
		},
		{
			name:      "InvalidRequest_InvalidStart",
			request:   "0410000000020400040002E0\r\n",
			readError: common.ErrInvalidPacket,
		},
		{
			name:      "InvalidRequest_NotOurAddress",
			request:   ":0510000000020400040002DF\r\n",
			readError: common.ErrNotOurAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			s, e := newModbusASCIIServerWithHandler(logger, port, 0x04, NewDefaultHandler(logger, 1024, 1024, 1024, 1024))
			assert.NoError(t, e)

			adu, err := s.acceptAndValidateTransaction()
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
				str := string(port.writeData)
				assert.Equal(t, tt.response, str)
				assert.Equal(t, tt.expectedRegisters, s.handler.(*DefaultHandler).HoldingRegisters[1:len(tt.expectedRegisters)+1])
			}
		})
	}
}

func TestGenerateLrc(t *testing.T) {
	data, err := hex.DecodeString("048102")
	assert.NoError(t, err)
	var lrc byte
	// then the data
	for _, b := range data {
		lrc += byte(b)
	}
	// Two's complement
	lrc = ^lrc + 1
	t.Logf("%x", []byte{lrc})
}
