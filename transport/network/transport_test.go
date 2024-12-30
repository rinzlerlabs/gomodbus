package network

import (
	"context"
	"encoding/hex"
	"io"
	"net"
	"testing"

	"github.com/rinzlerlabs/gomodbus/data"
	"github.com/rinzlerlabs/gomodbus/transport"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

type testConnection struct {
	readData  []byte
	writeData []byte
}

func (t *testConnection) Read(b []byte) (n int, err error) {
	if len(t.readData) == 0 {
		return 0, io.EOF
	}
	lenRead := copy(b, t.readData)
	t.readData = t.readData[lenRead:]
	return lenRead, nil
}

func (t *testConnection) Write(b []byte) (n int, err error) {
	t.writeData = b
	return len(b), nil
}

func (t *testConnection) Close() error {
	return nil
}

func (t *testConnection) RemoteAddr() net.Addr {
	return &addr{}
}

type addr struct{}

func (a *addr) Network() string {
	return "tcp"
}

func (a *addr) String() string {
	return "localhost:502"
}

func TestAcceptRequest(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()
	port := &testConnection{
		readData: []byte{0x00, 0x02, 0x00, 0x00, 0x00, 0x05, 0x01, 0x01, 0x00, 0x0A, 0x00, 0x0D},
	}
	tp := NewModbusServerTransport(port, logger)
	txn, err := tp.ReadRequest(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, txn)
	assert.Equal(t, []byte{0x00, 0x02}, txn.Header().(transport.NetworkHeader).TransactionID())
	assert.Equal(t, []byte{0x00, 0x00}, txn.Header().(transport.NetworkHeader).ProtocolID())
	assert.Equal(t, byte(0x01), txn.Header().(transport.NetworkHeader).UnitID())
	assert.Equal(t, data.FunctionCode(0x01), txn.PDU().FunctionCode())
	assert.Equal(t, []byte{0x00, 0x0A, 0x00, 0x0D}, data.ModbusOperationToBytes(txn.PDU().Operation()))
	assert.Equal(t, []byte{}, []byte(txn.Checksum()))
}

func TestReadCoilsRequest(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name      string
		request   string
		readError error
	}{
		{
			name:    "Valid",
			request: "0002000000050101000A000D",
		},
		// TODO: Add tests for bad headers
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			d, err := hex.DecodeString(tt.request)
			assert.NoError(t, err)
			port := &testConnection{
				readData: d,
			}
			tp := NewModbusServerTransport(port, logger)
			txn, err := tp.ReadRequest(ctx)
			if tt.readError != nil {
				assert.Error(t, err)
				assert.Nil(t, txn)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.Equal(t, []byte{0x00, 0x02}, txn.Header().(transport.NetworkHeader).TransactionID())
			assert.Equal(t, []byte{0x00, 0x00}, txn.Header().(transport.NetworkHeader).ProtocolID())
			assert.Equal(t, byte(0x01), txn.Header().(transport.NetworkHeader).UnitID())
			assert.Equal(t, data.ReadCoils, txn.PDU().FunctionCode())
			assert.Equal(t, uint16(0x0A), txn.PDU().Operation().(*data.ReadCoilsRequest).Offset())
			assert.Equal(t, int(0x0D), txn.PDU().Operation().(*data.ReadCoilsRequest).Count())
			assert.Equal(t, transport.ErrorCheck([]byte{}), txn.Checksum())
		})
	}
}

func TestReadDiscreteInputsRequest(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name      string
		request   string
		readError error
	}{
		{
			name:    "Valid",
			request: "0002000000050102000A000D",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			d, err := hex.DecodeString(tt.request)
			assert.NoError(t, err)
			port := &testConnection{
				readData: d,
			}
			tp := NewModbusServerTransport(port, logger)
			txn, err := tp.ReadRequest(ctx)
			if tt.readError != nil {
				assert.Error(t, err)
				assert.Nil(t, txn)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.Equal(t, []byte{0x00, 0x02}, txn.Header().(transport.NetworkHeader).TransactionID())
			assert.Equal(t, []byte{0x00, 0x00}, txn.Header().(transport.NetworkHeader).ProtocolID())
			assert.Equal(t, byte(0x01), txn.Header().(transport.NetworkHeader).UnitID())
			assert.Equal(t, data.ReadDiscreteInputs, txn.PDU().FunctionCode())
			assert.Equal(t, uint16(0x0A), txn.PDU().Operation().(*data.ReadDiscreteInputsRequest).Offset())
			assert.Equal(t, int(0x0D), txn.PDU().Operation().(*data.ReadDiscreteInputsRequest).Count())
			assert.Equal(t, transport.ErrorCheck([]byte{}), txn.Checksum())
		})
	}
}

func TestReadHoldingRegistersRequest(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name      string
		request   string
		readError error
	}{
		{
			name:    "Valid",
			request: "000200000005010300000002",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			d, err := hex.DecodeString(tt.request)
			assert.NoError(t, err)
			port := &testConnection{
				readData: d,
			}
			tp := NewModbusServerTransport(port, logger)
			txn, err := tp.ReadRequest(ctx)
			if tt.readError != nil {
				assert.Error(t, err)
				assert.Nil(t, txn)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.Equal(t, []byte{0x00, 0x02}, txn.Header().(transport.NetworkHeader).TransactionID())
			assert.Equal(t, []byte{0x00, 0x00}, txn.Header().(transport.NetworkHeader).ProtocolID())
			assert.Equal(t, byte(0x01), txn.Header().(transport.NetworkHeader).UnitID())
			assert.Equal(t, data.ReadHoldingRegisters, txn.PDU().FunctionCode())
			assert.Equal(t, uint16(0x00), txn.PDU().Operation().(*data.ReadHoldingRegistersRequest).Offset())
			assert.Equal(t, int(0x02), txn.PDU().Operation().(*data.ReadHoldingRegistersRequest).Count())
			assert.Equal(t, transport.ErrorCheck([]byte{}), txn.Checksum())
		})
	}
}

func TestReadInputRegistersRequest(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name      string
		request   string
		readError error
	}{
		{
			name:    "Valid",
			request: "000200000005010400000002",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			d, err := hex.DecodeString(tt.request)
			assert.NoError(t, err)
			port := &testConnection{
				readData: d,
			}
			tp := NewModbusServerTransport(port, logger)
			txn, err := tp.ReadRequest(ctx)
			if tt.readError != nil {
				assert.Error(t, err)
				assert.Nil(t, txn)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.Equal(t, []byte{0x00, 0x02}, txn.Header().(transport.NetworkHeader).TransactionID())
			assert.Equal(t, []byte{0x00, 0x00}, txn.Header().(transport.NetworkHeader).ProtocolID())
			assert.Equal(t, byte(0x01), txn.Header().(transport.NetworkHeader).UnitID())
			assert.Equal(t, data.ReadInputRegisters, txn.PDU().FunctionCode())
			assert.Equal(t, uint16(0x00), txn.PDU().Operation().(*data.ReadInputRegistersRequest).Offset())
			assert.Equal(t, int(0x02), txn.PDU().Operation().(*data.ReadInputRegistersRequest).Count())
			assert.Equal(t, transport.ErrorCheck([]byte{}), txn.Checksum())
		})
	}
}

func TestWriteSingleCoilRequest(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name      string
		request   string
		readError error
	}{
		{
			name:    "Valid",
			request: "0002000000050105000AFF00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			d, err := hex.DecodeString(tt.request)
			assert.NoError(t, err)
			port := &testConnection{
				readData: d,
			}
			tp := NewModbusServerTransport(port, logger)
			txn, err := tp.ReadRequest(ctx)
			if tt.readError != nil {
				assert.Error(t, err)
				assert.Nil(t, txn)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.Equal(t, []byte{0x00, 0x02}, txn.Header().(transport.NetworkHeader).TransactionID())
			assert.Equal(t, []byte{0x00, 0x00}, txn.Header().(transport.NetworkHeader).ProtocolID())
			assert.Equal(t, byte(0x01), txn.Header().(transport.NetworkHeader).UnitID())
			assert.Equal(t, data.WriteSingleCoil, txn.PDU().FunctionCode())
			assert.Equal(t, uint16(0x0A), txn.PDU().Operation().(*data.WriteSingleCoilRequest).Offset())
			assert.Equal(t, true, txn.PDU().Operation().(*data.WriteSingleCoilRequest).Value())
			assert.Equal(t, transport.ErrorCheck([]byte{}), txn.Checksum())
		})
	}
}

func TestWriteSingleRegisterRequest(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name      string
		request   string
		readError error
	}{
		{
			name:    "Valid",
			request: "000200000005010600100003",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			d, err := hex.DecodeString(tt.request)
			assert.NoError(t, err)
			port := &testConnection{
				readData: d,
			}
			tp := NewModbusServerTransport(port, logger)
			txn, err := tp.ReadRequest(ctx)
			if tt.readError != nil {
				assert.Error(t, err)
				assert.Nil(t, txn)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.Equal(t, []byte{0x00, 0x02}, txn.Header().(transport.NetworkHeader).TransactionID())
			assert.Equal(t, []byte{0x00, 0x00}, txn.Header().(transport.NetworkHeader).ProtocolID())
			assert.Equal(t, byte(0x01), txn.Header().(transport.NetworkHeader).UnitID())
			assert.Equal(t, data.WriteSingleRegister, txn.PDU().FunctionCode())
			assert.Equal(t, uint16(0x10), txn.PDU().Operation().(*data.WriteSingleRegisterRequest).Offset())
			assert.Equal(t, uint16(0x03), txn.PDU().Operation().(*data.WriteSingleRegisterRequest).Value())
			assert.Equal(t, transport.ErrorCheck([]byte{}), txn.Checksum())
		})
	}
}

func TestWriteMultipleCoilsRequest(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name      string
		request   string
		readError error
	}{
		{
			name:    "Valid",
			request: "000200000018010F0000001803018307",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			d, err := hex.DecodeString(tt.request)
			assert.NoError(t, err)
			port := &testConnection{
				readData: d,
			}
			tp := NewModbusServerTransport(port, logger)
			txn, err := tp.ReadRequest(ctx)
			if tt.readError != nil {
				assert.Error(t, err)
				assert.Nil(t, txn)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.Equal(t, []byte{0x00, 0x02}, txn.Header().(transport.NetworkHeader).TransactionID())
			assert.Equal(t, []byte{0x00, 0x00}, txn.Header().(transport.NetworkHeader).ProtocolID())
			assert.Equal(t, byte(0x01), txn.Header().(transport.NetworkHeader).UnitID())
			assert.Equal(t, data.WriteMultipleCoils, txn.PDU().FunctionCode())
			assert.Equal(t, uint16(0x00), txn.PDU().Operation().(*data.WriteMultipleCoilsRequest).Offset())
			assert.Equal(t, []bool{true, false, false, false, false, false, false, false, true, true, false, false, false, false, false, true, true, true, true, false, false, false, false, false}, txn.PDU().Operation().(*data.WriteMultipleCoilsRequest).Values())
			assert.Equal(t, transport.ErrorCheck([]byte{}), txn.Checksum())
		})
	}
}

func TestWriteMultipleRegistersRequest(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name      string
		request   string
		readError error
	}{
		{
			name:    "Valid",
			request: "0002000000200110000000020400040002",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			d, err := hex.DecodeString(tt.request)
			assert.NoError(t, err)
			port := &testConnection{
				readData: d,
			}
			tp := NewModbusServerTransport(port, logger)
			txn, err := tp.ReadRequest(ctx)
			if tt.readError != nil {
				assert.Error(t, err)
				assert.Nil(t, txn)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, txn)
			assert.Equal(t, []byte{0x00, 0x02}, txn.Header().(transport.NetworkHeader).TransactionID())
			assert.Equal(t, []byte{0x00, 0x00}, txn.Header().(transport.NetworkHeader).ProtocolID())
			assert.Equal(t, byte(0x01), txn.Header().(transport.NetworkHeader).UnitID())
			assert.Equal(t, data.WriteMultipleRegisters, txn.PDU().FunctionCode())
			assert.Equal(t, uint16(0x00), txn.PDU().Operation().(*data.WriteMultipleRegistersRequest).Offset())
			assert.Equal(t, []uint16{0x0004, 0x0002}, txn.PDU().Operation().(*data.WriteMultipleRegistersRequest).Values())
			assert.Equal(t, transport.ErrorCheck([]byte{}), txn.Checksum())
		})
	}
}
