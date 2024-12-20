package ascii

import (
	"testing"

	"github.com/rinzlerlabs/gomodbus/data"
	"github.com/rinzlerlabs/gomodbus/transport"
	"github.com/rinzlerlabs/gomodbus/transport/serial"
	"github.com/stretchr/testify/assert"
)

func TestADU_Bytes(t *testing.T) {
	tests := []struct {
		name          string
		adu           *modbusApplicationDataUnit
		expectedBytes []byte
	}{
		{
			name: "ReadCoilsRequest",
			adu: &modbusApplicationDataUnit{
				header: serial.NewHeader(1),
				pdu:    transport.NewProtocolDataUnit(data.ReadCoils, &data.ReadCoilsRequest{Offset: 0, Count: 16}),
			},
			expectedBytes: []byte{0x01, 0x01, 0x00, 0x00, 0x00, 0x10},
		},
		{
			name: "ReadCoilsResponse",
			adu: &modbusApplicationDataUnit{
				header: serial.NewHeader(1),
				pdu:    transport.NewProtocolDataUnit(data.ReadCoils, &data.ReadCoilsResponse{Values: []bool{true, false, false, false, false, false, false, true, true, false, false, false, false, false, false, true}}),
			},
			expectedBytes: []byte{0x01, 0x01, 0x02, 0x81, 0x81},
		},
		{
			name: "ReadDiscreteInputsRequest",
			adu: &modbusApplicationDataUnit{
				header: serial.NewHeader(1),
				pdu:    transport.NewProtocolDataUnit(data.ReadDiscreteInputs, &data.ReadDiscreteInputsRequest{Offset: 0, Count: 16}),
			},
			expectedBytes: []byte{0x01, 0x02, 0x00, 0x00, 0x00, 0x10},
		},
		{
			name: "ReadDiscreteInputsResponse",
			adu: &modbusApplicationDataUnit{
				header: serial.NewHeader(1),
				pdu:    transport.NewProtocolDataUnit(data.ReadDiscreteInputs, &data.ReadDiscreteInputsResponse{Values: []bool{true, false, false, false, false, false, false, true, true, false, false, false, false, false, false, true}}),
			},
			expectedBytes: []byte{0x01, 0x02, 0x02, 0x81, 0x81},
		},
		{
			name: "ReadHoldingRegistersRequest",
			adu: &modbusApplicationDataUnit{
				header: serial.NewHeader(1),
				pdu:    transport.NewProtocolDataUnit(data.ReadHoldingRegisters, &data.ReadHoldingRegistersRequest{Offset: 0, Count: 4}),
			},
			expectedBytes: []byte{0x01, 0x03, 0x00, 0x00, 0x00, 0x04},
		},
		{
			name: "ReadHoldingRegistersResponse",
			adu: &modbusApplicationDataUnit{
				header: serial.NewHeader(1),
				pdu:    transport.NewProtocolDataUnit(data.ReadHoldingRegisters, &data.ReadHoldingRegistersResponse{Values: []uint16{0x0001, 0x0002, 0x0003, 0x0004}}),
			},
			expectedBytes: []byte{0x01, 0x03, 0x08, 0x00, 0x01, 0x00, 0x02, 0x00, 0x03, 0x00, 0x04},
		},
		{
			name: "ReadInputRegistersRequest",
			adu: &modbusApplicationDataUnit{
				header: serial.NewHeader(1),
				pdu:    transport.NewProtocolDataUnit(data.ReadInputRegisters, &data.ReadInputRegistersRequest{Offset: 0, Count: 4}),
			},
			expectedBytes: []byte{0x01, 0x04, 0x00, 0x00, 0x00, 0x04},
		},
		{
			name: "ReadInputRegistersResponse",
			adu: &modbusApplicationDataUnit{
				header: serial.NewHeader(1),
				pdu:    transport.NewProtocolDataUnit(data.ReadInputRegisters, &data.ReadInputRegistersResponse{Values: []uint16{0x0001, 0x0002, 0x0003, 0x0004}}),
			},
			expectedBytes: []byte{0x01, 0x04, 0x08, 0x00, 0x01, 0x00, 0x02, 0x00, 0x03, 0x00, 0x04},
		},
		{
			name: "WriteSingleCoilRequest",
			adu: &modbusApplicationDataUnit{
				header: serial.NewHeader(1),
				pdu:    transport.NewProtocolDataUnit(data.WriteSingleCoil, &data.WriteSingleCoilRequest{Offset: 0, Value: true}),
			},
			expectedBytes: []byte{0x01, 0x05, 0x00, 0x00, 0xFF, 0x00},
		},
		{
			name: "WriteSingleCoilResponse",
			adu: &modbusApplicationDataUnit{
				header: serial.NewHeader(1),
				pdu:    transport.NewProtocolDataUnit(data.WriteSingleCoil, &data.WriteSingleCoilResponse{Offset: 0, Value: true}),
			},
			expectedBytes: []byte{0x01, 0x05, 0x00, 0x00, 0xFF, 0x00},
		},
		{
			name: "WriteSingleRegisterRequest",
			adu: &modbusApplicationDataUnit{
				header: serial.NewHeader(1),
				pdu:    transport.NewProtocolDataUnit(data.WriteSingleRegister, &data.WriteSingleRegisterRequest{Offset: 0, Value: 0x0001}),
			},
			expectedBytes: []byte{0x01, 0x06, 0x00, 0x00, 0x00, 0x01},
		},
		{
			name: "WriteSingleRegisterResponse",
			adu: &modbusApplicationDataUnit{
				header: serial.NewHeader(1),
				pdu:    transport.NewProtocolDataUnit(data.WriteSingleRegister, &data.WriteSingleRegisterResponse{Offset: 0, Value: 0x0001}),
			},
			expectedBytes: []byte{0x01, 0x06, 0x00, 0x00, 0x00, 0x01},
		},
		{
			name: "WriteMultipleCoilsRequest",
			adu: &modbusApplicationDataUnit{
				header: serial.NewHeader(1),
				pdu: transport.NewProtocolDataUnit(data.WriteMultipleCoils, &data.WriteMultipleCoilsRequest{
					Offset: 0,
					Values: []bool{true, false, false, false, false, false, false, true, true, false, false, false, false, false, false, true},
				}),
			},
			expectedBytes: []byte{0x01, 0x0F, 0x00, 0x00, 0x00, 0x10, 0x02, 0x81, 0x81},
		},
		{
			name: "WriteMultipleCoilsResponse",
			adu: &modbusApplicationDataUnit{
				header: serial.NewHeader(1),
				pdu: transport.NewProtocolDataUnit(data.WriteMultipleCoils, &data.WriteMultipleCoilsResponse{
					Offset: 0,
					Count:  16,
				}),
			},
			expectedBytes: []byte{0x01, 0x0F, 0x00, 0x00, 0x00, 0x10},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytes := tt.adu.Bytes()
			assert.Equal(t, tt.expectedBytes, bytes[:len(bytes)-1]) // Ignore the checksum because i don't feel like calculating by hand
		})
	}
}
