package server

import (
	"testing"

	"github.com/rinzlerlabs/gomodbus/data"
	"github.com/rinzlerlabs/gomodbus/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zapcore"
)

type mockADU struct {
	mock.Mock
}

func (m *mockADU) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	args := m.Called(encoder)
	return args.Error(0)
}

func (m *mockADU) Bytes() []byte {
	args := m.Called()
	return args.Get(0).([]byte)
}

func (m *mockADU) Header() transport.Header {
	args := m.Called()
	return args.Get(0).(transport.Header)
}
func (m *mockADU) PDU() *transport.ProtocolDataUnit {
	args := m.Called()
	return args.Get(0).(*transport.ProtocolDataUnit)
}
func (m *mockADU) Checksum() transport.ErrorCheck {
	args := m.Called()
	return args.Get(0).(transport.ErrorCheck)
}

func TestServerStatsTrackServerStats(t *testing.T) {
	tests := []struct {
		name                        string
		ops                         []data.ModbusOperation
		readCoilsCount              uint64
		readDiscreteInputsCount     uint64
		readHoldingRegistersCount   uint64
		readInputRegistersCount     uint64
		writeSingleCoilCount        uint64
		writeSingleRegisterCount    uint64
		writeMultipleCoilsCount     uint64
		writeMultipleRegistersCount uint64
		expectedTotalCount          uint64
	}{
		{
			name:               "ReadCoils_Once",
			ops:                []data.ModbusOperation{&data.ReadCoilsRequest{}},
			readCoilsCount:     1,
			expectedTotalCount: 1,
		},
		{
			name:               "ReadCoils_Twice",
			ops:                []data.ModbusOperation{&data.ReadCoilsRequest{}, &data.ReadCoilsRequest{}},
			readCoilsCount:     2,
			expectedTotalCount: 2,
		},
		{
			name:               "ReadCoils_None",
			ops:                []data.ModbusOperation{},
			readCoilsCount:     0,
			expectedTotalCount: 0,
		},
		{
			name:                    "ReadDiscreteInputs_Once",
			ops:                     []data.ModbusOperation{&data.ReadDiscreteInputsRequest{}},
			readDiscreteInputsCount: 1,
			expectedTotalCount:      1,
		},
		{
			name:                    "ReadDiscreteInputs_Twice",
			ops:                     []data.ModbusOperation{&data.ReadDiscreteInputsRequest{}, &data.ReadDiscreteInputsRequest{}},
			readDiscreteInputsCount: 2,
			expectedTotalCount:      2,
		},
		{
			name:                    "ReadDiscreteInputs_None",
			ops:                     []data.ModbusOperation{},
			readDiscreteInputsCount: 0,
			expectedTotalCount:      0,
		},
		{
			name:                      "ReadHoldingRegisters_Once",
			ops:                       []data.ModbusOperation{&data.ReadHoldingRegistersRequest{}},
			readHoldingRegistersCount: 1,
			expectedTotalCount:        1,
		},
		{
			name:                      "ReadHoldingRegisters_Twice",
			ops:                       []data.ModbusOperation{&data.ReadHoldingRegistersRequest{}, &data.ReadHoldingRegistersRequest{}},
			readHoldingRegistersCount: 2,
			expectedTotalCount:        2,
		},
		{
			name:                      "ReadHoldingRegisters_None",
			ops:                       []data.ModbusOperation{},
			readHoldingRegistersCount: 0,
			expectedTotalCount:        0,
		},
		{
			name:                    "ReadInputRegisters_Once",
			ops:                     []data.ModbusOperation{&data.ReadInputRegistersRequest{}},
			readInputRegistersCount: 1,
			expectedTotalCount:      1,
		},
		{
			name:                    "ReadInputRegisters_Twice",
			ops:                     []data.ModbusOperation{&data.ReadInputRegistersRequest{}, &data.ReadInputRegistersRequest{}},
			readInputRegistersCount: 2,
			expectedTotalCount:      2,
		},
		{
			name:                    "ReadInputRegisters_None",
			ops:                     []data.ModbusOperation{},
			readInputRegistersCount: 0,
			expectedTotalCount:      0,
		},
		{
			name:                 "WriteSingleCoil_Once",
			ops:                  []data.ModbusOperation{&data.WriteSingleCoilRequest{}},
			writeSingleCoilCount: 1,
			expectedTotalCount:   1,
		},
		{
			name:                 "WriteSingleCoil_Twice",
			ops:                  []data.ModbusOperation{&data.WriteSingleCoilRequest{}, &data.WriteSingleCoilRequest{}},
			writeSingleCoilCount: 2,
			expectedTotalCount:   2,
		},
		{
			name:                 "WriteSingleCoil_None",
			ops:                  []data.ModbusOperation{},
			writeSingleCoilCount: 0,
			expectedTotalCount:   0,
		},
		{
			name:                     "WriteSingleRegister_Once",
			ops:                      []data.ModbusOperation{&data.WriteSingleRegisterRequest{}},
			writeSingleRegisterCount: 1,
			expectedTotalCount:       1,
		},
		{
			name:                     "WriteSingleRegister_Twice",
			ops:                      []data.ModbusOperation{&data.WriteSingleRegisterRequest{}, &data.WriteSingleRegisterRequest{}},
			writeSingleRegisterCount: 2,
			expectedTotalCount:       2,
		},
		{
			name:                     "WriteSingleRegister_None",
			ops:                      []data.ModbusOperation{},
			writeSingleRegisterCount: 0,
			expectedTotalCount:       0,
		},
		{
			name:                    "WriteMultipleCoils_Once",
			ops:                     []data.ModbusOperation{&data.WriteMultipleCoilsRequest{}},
			writeMultipleCoilsCount: 1,
			expectedTotalCount:      1,
		},
		{
			name:                    "WriteMultipleCoils_Twice",
			ops:                     []data.ModbusOperation{&data.WriteMultipleCoilsRequest{}, &data.WriteMultipleCoilsRequest{}},
			writeMultipleCoilsCount: 2,
			expectedTotalCount:      2,
		},
		{
			name:                    "WriteMultipleCoils_None",
			ops:                     []data.ModbusOperation{},
			writeMultipleCoilsCount: 0,
			expectedTotalCount:      0,
		},
		{
			name:                        "WriteMultipleRegisters_Once",
			ops:                         []data.ModbusOperation{&data.WriteMultipleRegistersRequest{}},
			writeMultipleRegistersCount: 1,
			expectedTotalCount:          1,
		},
		{
			name:                        "WriteMultipleRegisters_Twice",
			ops:                         []data.ModbusOperation{&data.WriteMultipleRegistersRequest{}, &data.WriteMultipleRegistersRequest{}},
			writeMultipleRegistersCount: 2,
			expectedTotalCount:          2,
		},
		{
			name:                        "WriteMultipleRegisters_None",
			ops:                         []data.ModbusOperation{},
			writeMultipleRegistersCount: 0,
			expectedTotalCount:          0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := NewServerStats()
			assert.NotNil(t, stats)

			for _, op := range tt.ops {
				mockADU := &mockADU{}
				mockADU.On("PDU").Return(transport.NewProtocolDataUnit(op)).Once()
				stats.AddRequest(mockADU)
			}

			assert.Equal(t, tt.readCoilsCount, stats.TotalReadCoilsRequests)
			assert.Equal(t, tt.readDiscreteInputsCount, stats.TotalReadDiscreteInputsRequests)
			assert.Equal(t, tt.readHoldingRegistersCount, stats.TotalReadHoldingRegistersRequests)
			assert.Equal(t, tt.readInputRegistersCount, stats.TotalReadInputRegistersRequests)
			assert.Equal(t, tt.writeSingleCoilCount, stats.TotalWriteSingleCoilRequests)
			assert.Equal(t, tt.writeSingleRegisterCount, stats.TotalWriteSingleRegisterRequests)
			assert.Equal(t, tt.writeMultipleCoilsCount, stats.TotalWriteMultipleCoilsRequests)
			assert.Equal(t, tt.writeMultipleRegistersCount, stats.TotalWriteMultipleRegistersRequests)
			assert.Equal(t, tt.expectedTotalCount, stats.TotalRequests)

			m := stats.AsMap()
			assert.Equal(t, tt.readCoilsCount, m["TotalReadCoilsRequests"])
			assert.Equal(t, tt.readDiscreteInputsCount, m["TotalReadDiscreteInputsRequests"])
			assert.Equal(t, tt.readHoldingRegistersCount, m["TotalReadHoldingRegistersRequests"])
			assert.Equal(t, tt.readInputRegistersCount, m["TotalReadInputRegistersRequests"])
			assert.Equal(t, tt.writeSingleCoilCount, m["TotalWriteSingleCoilRequests"])
			assert.Equal(t, tt.writeSingleRegisterCount, m["TotalWriteSingleRegisterRequests"])
			assert.Equal(t, tt.writeMultipleCoilsCount, m["TotalWriteMultipleCoilsRequests"])
			assert.Equal(t, tt.writeMultipleRegistersCount, m["TotalWriteMultipleRegistersRequests"])
			assert.Equal(t, tt.expectedTotalCount, m["TotalRequests"])
		})
	}
}
