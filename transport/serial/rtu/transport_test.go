package rtu

import (
	"context"
	"encoding/hex"
	"errors"
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
		readData: []byte{0x04, 0x01, 0x00, 0x0A, 0x00, 0x0D, 0xDD, 0x98},
	}
	tp := NewModbusServerTransport(port, logger, 0x04)
	txn, err := tp.AcceptRequest(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, txn)
	adu := txn.Frame()
	assert.Equal(t, uint16(0x04), adu.Header().(transport.SerialHeader).Address())
	assert.Equal(t, data.ReadCoils, adu.PDU().FunctionCode())
	assert.Equal(t, []byte{0x00, 0x0A, 0x00, 0x0D}, adu.PDU().Operation().Bytes())
	assert.Equal(t, []byte{0xDD, 0x98}, []byte(adu.Checksum()))
}

func TestReadCoils(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name      string
		request   []byte
		readError error
	}{
		{
			name:    "Valid",
			request: []byte{0x04, 0x01, 0x00, 0x0A, 0x00, 0x0D, 0xDD, 0x98},
		},
		{
			name:      "InvalidRequest_IvalidChecksum",
			request:   []byte{0x04, 0x01, 0x00, 0x0A, 0x00, 0x0D, 0xDD, 0x99},
			readError: common.ErrInvalidChecksum,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			tp := NewModbusServerTransport(port, logger, 0x04)
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
			assert.Equal(t, uint16(0x0A), txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.ReadCoilsRequest).Offset)
			assert.Equal(t, uint16(0x0D), txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.ReadCoilsRequest).Count)
			assert.Equal(t, transport.ErrorCheck([]byte{0xDD, 0x98}), txn.Frame().ApplicationDataUnit.Checksum())
		})
	}
}

func TestReadDiscreteInputs(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name      string
		request   []byte
		readError error
	}{
		{
			name:    "Valid",
			request: []byte{0x04, 0x02, 0x00, 0x0A, 0x00, 0x0D, 0x99, 0x98},
		},
		{
			name:      "InvalidRequest_IvalidChecksum",
			request:   []byte{0x04, 0x02, 0x00, 0x0A, 0x00, 0x0D, 0x99, 0x99},
			readError: common.ErrInvalidChecksum,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			tp := NewModbusServerTransport(port, logger, 0x04)
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
			assert.Equal(t, uint16(0x0A), txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.ReadDiscreteInputsRequest).Offset)
			assert.Equal(t, uint16(0x0D), txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.ReadDiscreteInputsRequest).Count)
			assert.Equal(t, transport.ErrorCheck([]byte{0x99, 0x98}), txn.Frame().ApplicationDataUnit.Checksum())
		})
	}
}

func TestReadHoldingRegisters(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name      string
		request   []byte
		readError error
	}{
		{
			name:    "Valid",
			request: []byte{0x04, 0x03, 0x00, 0x00, 0x00, 0x02, 0xC4, 0x5E},
		},
		{
			name:      "InvalidRequest_IvalidChecksum",
			request:   []byte{0x04, 0x03, 0x00, 0x00, 0x00, 0x02, 0xC4, 0x5F},
			readError: common.ErrInvalidChecksum,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			tp := NewModbusServerTransport(port, logger, 0x04)
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
			assert.Equal(t, uint16(0x00), txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.ReadHoldingRegistersRequest).Offset)
			assert.Equal(t, uint16(0x02), txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.ReadHoldingRegistersRequest).Count)
			assert.Equal(t, transport.ErrorCheck([]byte{0xC4, 0x5E}), txn.Frame().ApplicationDataUnit.Checksum())
		})
	}
}

func TestReadInputRegisters(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name      string
		request   []byte
		readError error
	}{
		{
			name:    "Valid",
			request: []byte{0x04, 0x04, 0x00, 0x00, 0x00, 0x02, 0x71, 0x9E},
		},
		{
			name:      "InvalidRequest_IvalidChecksum",
			request:   []byte{0x04, 0x04, 0x00, 0x00, 0x00, 0x02, 0x71, 0x9F},
			readError: common.ErrInvalidChecksum,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			tp := NewModbusServerTransport(port, logger, 0x04)
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
			assert.Equal(t, uint16(0x00), txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.ReadInputRegistersRequest).Offset)
			assert.Equal(t, uint16(0x02), txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.ReadInputRegistersRequest).Count)
			assert.Equal(t, transport.ErrorCheck([]byte{0x71, 0x9E}), txn.Frame().ApplicationDataUnit.Checksum())
		})
	}
}

func TestWriteSingleCoil(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name      string
		request   []byte
		readError error
	}{
		{
			name:    "Valid",
			request: []byte{0x04, 0x05, 0x00, 0x0A, 0xFF, 0x00, 0xAC, 0x6D},
		},
		{
			name:      "InvalidRequest_IvalidChecksum",
			request:   []byte{0x04, 0x05, 0x00, 0x0A, 0xFF, 0x00, 0xAC, 0x6E},
			readError: common.ErrInvalidChecksum,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			tp := NewModbusServerTransport(port, logger, 0x04)
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
			assert.Equal(t, uint16(0x0A), txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.WriteSingleCoilRequest).Offset)
			assert.Equal(t, true, txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.WriteSingleCoilRequest).Value)
			assert.Equal(t, transport.ErrorCheck([]byte{0xAC, 0x6D}), txn.Frame().ApplicationDataUnit.Checksum())
		})
	}
}

func TestWriteSingleRegister(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name      string
		request   []byte
		readError error
	}{
		{
			name:    "Valid",
			request: []byte{0x04, 0x06, 0x00, 0x10, 0x00, 0x03, 0xC8, 0x5B},
		},
		{
			name:      "InvalidRequest_IvalidChecksum",
			request:   []byte{0x04, 0x06, 0x00, 0x10, 0x00, 0x03, 0xC8, 0x5C},
			readError: common.ErrInvalidChecksum,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			tp := NewModbusServerTransport(port, logger, 0x04)
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
			assert.Equal(t, uint16(0x10), txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.WriteSingleRegisterRequest).Offset)
			assert.Equal(t, uint16(0x03), txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.WriteSingleRegisterRequest).Value)
			assert.Equal(t, transport.ErrorCheck([]byte{0xC8, 0x5B}), txn.Frame().ApplicationDataUnit.Checksum())
		})
	}
}

func TestWriteMultipleCoils(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name      string
		request   []byte
		readError error
	}{
		{
			name:    "Valid",
			request: []byte{0x04, 0x0F, 0x00, 0x00, 0x00, 0x18, 0x03, 0x01, 0x83, 0x07, 0x21, 0x56},
		},
		{
			name:      "InvalidRequest_IvalidChecksum",
			request:   []byte{0x04, 0x0F, 0x00, 0x00, 0x00, 0x18, 0x03, 0x01, 0x83, 0x07, 0x21, 0x57},
			readError: common.ErrInvalidChecksum,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			tp := NewModbusServerTransport(port, logger, 0x04)
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
			assert.Equal(t, uint16(0x00), txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.WriteMultipleCoilsRequest).Offset)
			assert.Equal(t, []bool{true, false, false, false, false, false, false, false, true, true, false, false, false, false, false, true, true, true, true, false, false, false, false, false}, txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.WriteMultipleCoilsRequest).Values)
			assert.Equal(t, transport.ErrorCheck([]byte{0x21, 0x56}), txn.Frame().ApplicationDataUnit.Checksum())
		})
	}
}

func TestWriteMultipleRegisters(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name      string
		request   []byte
		readError error
	}{
		{
			name:    "Valid",
			request: []byte{0x04, 0x10, 0x00, 0x00, 0x00, 0x02, 0x04, 0x00, 0x04, 0x00, 0x02, 0x22, 0x63},
		},
		{
			name:      "InvalidRequest_InvalidChecksum",
			request:   []byte{0x04, 0x10, 0x00, 0x00, 0x00, 0x02, 0x04, 0x00, 0x04, 0x00, 0x02, 0x22, 0x64},
			readError: common.ErrInvalidChecksum,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			tp := NewModbusServerTransport(port, logger, 0x04)
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
			assert.Equal(t, uint16(0x00), txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.WriteMultipleRegistersRequest).Offset)
			assert.Equal(t, []uint16{0x0004, 0x0002}, txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.WriteMultipleRegistersRequest).Values)
			assert.Equal(t, transport.ErrorCheck([]byte{0x22, 0x63}), txn.Frame().ApplicationDataUnit.Checksum())
		})
	}
}

func TestReadCoils_DisjoinedReads(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name      string
		requests  [][]byte
		readError error
	}{
		{
			name:     "SingleRead",
			requests: [][]byte{{0x04, 0x01, 0x00, 0x0A, 0x00, 0x0D, 0xDD, 0x98}},
		},
		{
			name:     "OneBytePerRead",
			requests: [][]byte{{0x04}, {0x01}, {0x00}, {0x0A}, {0x00}, {0x0D}, {0xDD}, {0x98}},
		},
		{
			name:     "TwoBytesPerRead",
			requests: [][]byte{{0x04, 0x01}, {0x00, 0x0A}, {0x00, 0x0D}, {0xDD, 0x98}},
		},
		{
			name:     "ThreeBytesPerRead",
			requests: [][]byte{{0x04, 0x01, 0x00}, {0x0A, 0x00, 0x0D}, {0xDD, 0x98}},
		},
		{
			name:     "FourBytesPerRead",
			requests: [][]byte{{0x04, 0x01, 0x00, 0x0A}, {0x00, 0x0D, 0xDD, 0x98}},
		},
		{
			name:     "ExtraBytesOnTheEnd",
			requests: [][]byte{{0x04, 0x01, 0x00, 0x0A, 0x00, 0x0D, 0xDD}, {0x98, 0x00, 0x00}},
		},
		{
			name:     "ExtraBytesAtTheStart",
			requests: [][]byte{{0x03, 0x00}, {0x04, 0x01, 0x00, 0x0A, 0x00, 0x0D, 0xDD}, {0x98, 0x00, 0x00}},
		},
		{
			name:     "ExtraBytesAtTheStart2",
			requests: [][]byte{{0x03, 0x00, 0x04, 0x01, 0x00, 0x0A, 0x00, 0x0D, 0xDD}, {0x98, 0x00, 0x00}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			port := &rtuSerialPort{
				reads: tt.requests,
			}
			tp := NewModbusServerTransport(port, logger, 0x04)
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
			assert.Equal(t, uint16(0x0A), txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.ReadCoilsRequest).Offset)
			assert.Equal(t, uint16(0x0D), txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.ReadCoilsRequest).Count)
			assert.Equal(t, transport.ErrorCheck([]byte{0xDD, 0x98}), txn.Frame().ApplicationDataUnit.Checksum())
		})
	}
}

func TestBigBlobOfBytes(t *testing.T) {
	logger := zaptest.NewLogger(t)
	request, err := hex.DecodeString("5B10008C00FFA700510000003B001E001E0000000000000000DA445B10008C000D1A000000000003004E00510000003B001E001E0000000000000000DA445B10008C000D1A000000000003004E00510000003B001E001E0000000000000000DA445B10008C000D1A000000000003004E00510000003B001E001E0000000000000000DA445B10008C000D1A000000000003004E00510000003B001E001E0000000000000000DA445B10008C000D1A000000000003004E00510000003B001E001E00")
	assert.NoError(t, err)
	ctx := context.Background()
	port := &rtuSerialPort{
		reads: [][]byte{request},
	}
	tp := NewModbusServerTransport(port, logger, 0x5B)
	txn, err := tp.AcceptRequest(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, txn)
	assert.Equal(t, uint16(0x5B), txn.Frame().ApplicationDataUnit.Header().(transport.SerialHeader).Address())
	assert.Equal(t, data.WriteMultipleRegisters, txn.Frame().ApplicationDataUnit.PDU().FunctionCode())
	assert.Equal(t, uint16(0x008C), txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.WriteMultipleRegistersRequest).Offset)
	assert.Equal(t, 13, len(txn.Frame().ApplicationDataUnit.PDU().Operation().(*data.WriteMultipleRegistersRequest).Values))

	txn, err = tp.AcceptRequest(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, txn)

	txn, err = tp.AcceptRequest(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, txn)

	txn, err = tp.AcceptRequest(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, txn)

	txn, err = tp.AcceptRequest(ctx)
	assert.Error(t, err)
	assert.Nil(t, txn)
}

type rtuSerialPort struct {
	reads  [][]byte
	writes [][]byte
}

func (t *rtuSerialPort) Read(b []byte) (n int, err error) {
	if len(t.reads) == 0 {
		return 0, io.EOF
	}
	lenRead := copy(b, t.reads[0])
	if lenRead < len(t.reads[0]) {
		t.reads[0] = t.reads[0][lenRead:]
		return lenRead, nil
	} else if lenRead == len(t.reads[0]) {
		t.reads = t.reads[1:]
		return lenRead, nil
	} else {
		return 0, errors.New("read too much")
	}
}

func (t *rtuSerialPort) Write(b []byte) (n int, err error) {
	t.writes = append(t.writes, b)
	return len(b), nil
}

func (t *rtuSerialPort) Close() error {
	return nil
}
