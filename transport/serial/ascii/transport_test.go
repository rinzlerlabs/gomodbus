package ascii

import (
	"context"
	"io"
	"testing"

	"github.com/rinzlerlabs/gomodbus/common"
	"github.com/rinzlerlabs/gomodbus/data"
	"github.com/rinzlerlabs/gomodbus/transport"
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

func TestAcceptRequest(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()
	port := &testSerialPort{
		readData: []byte{0x3A, 0x30, 0x32, 0x30, 0x31, 0x30, 0x30, 0x32, 0x30, 0x30, 0x30, 0x30, 0x43, 0x44, 0x31, 0x0D, 0x0A},
	}
	tp := NewModbusTransport(port, logger)
	txn, err := tp.AcceptRequest(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, txn)
	adu := txn.Frame()
	assert.Equal(t, uint16(0x02), adu.Header().(transport.SerialHeader).Address())
	assert.Equal(t, data.FunctionCode(0x01), adu.PDU().FunctionCode())
	assert.Equal(t, []byte{0x00, 0x20, 0x00, 0xC}, data.ModbusOperationToBytes(adu.PDU().Operation()))
	assert.Equal(t, []byte{0xD1}, []byte(adu.Checksum()))
}

func TestReadCoils(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name      string
		request   string
		readError error
	}{
		{
			name:    "Valid",
			request: ":0401000A000DE4\r\n",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			tp := NewModbusTransport(port, logger)
			txn, err := tp.AcceptRequest(ctx)
			if tt.readError != nil {
				assert.Error(t, err)
				assert.Nil(t, txn)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.Equal(t, uint16(0x04), txn.Frame().ApplicationDataUnit.Header().(transport.SerialHeader).Address())
			assert.Equal(t, data.ReadCoils, txn.Frame().ApplicationDataUnit.PDU().FunctionCode())
			assert.Equal(t, uint16(0x0A), txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.ReadCoilsRequest).Offset())
			assert.Equal(t, int(0x0D), txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.ReadCoilsRequest).Count())
			assert.Equal(t, transport.ErrorCheck([]byte{0xE4}), txn.Frame().ApplicationDataUnit.Checksum())
		})
	}
}

func TestReadDiscreteInputs(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name      string
		request   string
		readError error
	}{
		{
			name:    "Valid",
			request: ":0402000A000DE3\r\n",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			tp := NewModbusTransport(port, logger)
			txn, err := tp.AcceptRequest(ctx)
			if tt.readError != nil {
				assert.Error(t, err)
				assert.Nil(t, txn)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.Equal(t, uint16(0x04), txn.Frame().ApplicationDataUnit.Header().(transport.SerialHeader).Address())
			assert.Equal(t, data.ReadDiscreteInputs, txn.Frame().ApplicationDataUnit.PDU().FunctionCode())
			assert.Equal(t, uint16(0x0A), txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.ReadDiscreteInputsRequest).Offset())
			assert.Equal(t, int(0x0D), txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.ReadDiscreteInputsRequest).Count())
			assert.Equal(t, transport.ErrorCheck([]byte{0xE3}), txn.Frame().ApplicationDataUnit.Checksum())
		})
	}
}

func TestReadHoldingRegisters(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name      string
		request   string
		readError error
	}{
		{
			name:    "Valid",
			request: ":040300000002F7\r\n",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			tp := NewModbusTransport(port, logger)
			txn, err := tp.AcceptRequest(ctx)
			if tt.readError != nil {
				assert.Error(t, err)
				assert.Nil(t, txn)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.Equal(t, uint16(0x04), txn.Frame().ApplicationDataUnit.Header().(transport.SerialHeader).Address())
			assert.Equal(t, data.ReadHoldingRegisters, txn.Frame().ApplicationDataUnit.PDU().FunctionCode())
			assert.Equal(t, uint16(0x00), txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.ReadHoldingRegistersRequest).Offset())
			assert.Equal(t, int(0x02), txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.ReadHoldingRegistersRequest).Count())
			assert.Equal(t, transport.ErrorCheck([]byte{0xF7}), txn.Frame().ApplicationDataUnit.Checksum())
		})
	}
}

func TestReadInputRegisters(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name      string
		request   string
		readError error
	}{
		{
			name:    "Valid",
			request: ":040400000002F6\r\n",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			tp := NewModbusTransport(port, logger)
			txn, err := tp.AcceptRequest(ctx)
			if tt.readError != nil {
				assert.Error(t, err)
				assert.Nil(t, txn)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.Equal(t, uint16(0x04), txn.Frame().ApplicationDataUnit.Header().(transport.SerialHeader).Address())
			assert.Equal(t, data.ReadInputRegisters, txn.Frame().ApplicationDataUnit.PDU().FunctionCode())
			assert.Equal(t, uint16(0x00), txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.ReadInputRegistersRequest).Offset())
			assert.Equal(t, int(0x02), txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.ReadInputRegistersRequest).Count())
			assert.Equal(t, transport.ErrorCheck([]byte{0xF6}), txn.Frame().ApplicationDataUnit.Checksum())
		})
	}
}

func TestWriteSingleCoil(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name      string
		request   string
		readError error
	}{
		{
			name:    "Valid",
			request: ":0405000AFF00EE\r\n",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			tp := NewModbusTransport(port, logger)
			txn, err := tp.AcceptRequest(ctx)
			if tt.readError != nil {
				assert.Error(t, err)
				assert.Nil(t, txn)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.Equal(t, uint16(0x04), txn.Frame().ApplicationDataUnit.Header().(transport.SerialHeader).Address())
			assert.Equal(t, data.WriteSingleCoil, txn.Frame().ApplicationDataUnit.PDU().FunctionCode())
			assert.Equal(t, uint16(0x0A), txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.WriteSingleCoilRequest).Offset())
			assert.Equal(t, true, txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.WriteSingleCoilRequest).Value())
			assert.Equal(t, transport.ErrorCheck([]byte{0xEE}), txn.Frame().ApplicationDataUnit.Checksum())
		})
	}
}

func TestWriteSingleRegister(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name      string
		request   string
		readError error
	}{
		{
			name:    "Valid",
			request: ":040600100003E3\r\n",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			tp := NewModbusTransport(port, logger)
			txn, err := tp.AcceptRequest(ctx)
			if tt.readError != nil {
				assert.Error(t, err)
				assert.Nil(t, txn)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.Equal(t, uint16(0x04), txn.Frame().ApplicationDataUnit.Header().(transport.SerialHeader).Address())
			assert.Equal(t, data.WriteSingleRegister, txn.Frame().ApplicationDataUnit.PDU().FunctionCode())
			assert.Equal(t, uint16(0x10), txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.WriteSingleRegisterRequest).Offset())
			assert.Equal(t, uint16(0x03), txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.WriteSingleRegisterRequest).Value())
			assert.Equal(t, transport.ErrorCheck([]byte{0xE3}), txn.Frame().ApplicationDataUnit.Checksum())
		})
	}
}

func TestWriteMultipleCoils(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name      string
		request   string
		readError error
	}{
		{
			name:    "Valid",
			request: ":040F000000180301830747\r\n",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			tp := NewModbusTransport(port, logger)
			txn, err := tp.AcceptRequest(ctx)
			if tt.readError != nil {
				assert.Error(t, err)
				assert.Nil(t, txn)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.Equal(t, uint16(0x04), txn.Frame().ApplicationDataUnit.Header().(transport.SerialHeader).Address())
			assert.Equal(t, data.WriteMultipleCoils, txn.Frame().ApplicationDataUnit.PDU().FunctionCode())
			assert.Equal(t, uint16(0x00), txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.WriteMultipleCoilsRequest).Offset())
			assert.Equal(t, []bool{true, false, false, false, false, false, false, false, true, true, false, false, false, false, false, true, true, true, true, false, false, false, false, false}, txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.WriteMultipleCoilsRequest).Values())
			assert.Equal(t, transport.ErrorCheck([]byte{0x47}), txn.Frame().ApplicationDataUnit.Checksum())
		})
	}
}

func TestWriteMultipleRegisters(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name      string
		request   string
		readError error
	}{
		{
			name:    "Valid",
			request: ":0410000000020400040002E0\r\n",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			tp := NewModbusTransport(port, logger)
			txn, err := tp.AcceptRequest(ctx)
			if tt.readError != nil {
				assert.Error(t, err)
				assert.Nil(t, txn)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.Equal(t, uint16(0x04), txn.Frame().ApplicationDataUnit.Header().(transport.SerialHeader).Address())
			assert.Equal(t, data.WriteMultipleRegisters, txn.Frame().ApplicationDataUnit.PDU().FunctionCode())
			assert.Equal(t, uint16(0x00), txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.WriteMultipleRegistersRequest).Offset())
			assert.Equal(t, []uint16{0x0004, 0x0002}, txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.WriteMultipleRegistersRequest).Values())
			assert.Equal(t, transport.ErrorCheck([]byte{0xE0}), txn.Frame().ApplicationDataUnit.Checksum())
		})
	}
}
