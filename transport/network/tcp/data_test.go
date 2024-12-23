package tcp

import (
	"testing"

	"github.com/rinzlerlabs/gomodbus/data"
	"github.com/rinzlerlabs/gomodbus/transport"
	"github.com/rinzlerlabs/gomodbus/transport/network"
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
				header: network.NewHeader([]byte{0x00, 0x01}, []byte{0x00, 0x00}, 0x01),
				pdu:    transport.NewProtocolDataUnit(&data.ReadCoilsRequest{Offset: 0, Count: 16}),
			},
			expectedBytes: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x05, 0x01, 0x01, 0x00, 0x00, 0x00, 0x10},
		},
		{
			name: "ReadCoilsResponse",
			adu: &modbusApplicationDataUnit{
				header: network.NewHeader([]byte{0x00, 0x01}, []byte{0x00, 0x00}, 0x01),
				pdu:    transport.NewProtocolDataUnit(&data.ReadCoilsResponse{Values: []bool{true, false, false, false, false, false, false, true, true, false, false, false, false, false, false, true}}),
			},
			expectedBytes: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x04, 0x01, 0x01, 0x02, 0x81, 0x81},
		},
		{
			name: "ReadDiscreteInputsRequest",
			adu: &modbusApplicationDataUnit{
				header: network.NewHeader([]byte{0x00, 0x01}, []byte{0x00, 0x00}, 0x01),
				pdu:    transport.NewProtocolDataUnit(&data.ReadDiscreteInputsRequest{Offset: 0, Count: 16}),
			},
			expectedBytes: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x05, 0x01, 0x02, 0x00, 0x00, 0x00, 0x10},
		},
		{
			name: "ReadDiscreteInputsResponse",
			adu: &modbusApplicationDataUnit{
				header: network.NewHeader([]byte{0x00, 0x01}, []byte{0x00, 0x00}, 0x01),
				pdu:    transport.NewProtocolDataUnit(&data.ReadDiscreteInputsResponse{Values: []bool{true, false, false, false, false, false, false, true, true, false, false, false, false, false, false, true}}),
			},
			expectedBytes: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x04, 0x01, 0x02, 0x02, 0x81, 0x81},
		},
		{
			name: "ReadHoldingRegistersRequest",
			adu: &modbusApplicationDataUnit{
				header: network.NewHeader([]byte{0x00, 0x01}, []byte{0x00, 0x00}, 0x01),
				pdu:    transport.NewProtocolDataUnit(&data.ReadHoldingRegistersRequest{Offset: 0, Count: 4}),
			},
			expectedBytes: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x05, 0x01, 0x03, 0x00, 0x00, 0x00, 0x04},
		},
		{
			name: "ReadHoldingRegistersResponse",
			adu: &modbusApplicationDataUnit{
				header: network.NewHeader([]byte{0x00, 0x01}, []byte{0x00, 0x00}, 0x01),
				pdu:    transport.NewProtocolDataUnit(&data.ReadHoldingRegistersResponse{Values: []uint16{0x0001, 0x0002, 0x0003, 0x0004}}),
			},
			expectedBytes: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x0A, 0x01, 0x03, 0x08, 0x00, 0x01, 0x00, 0x02, 0x00, 0x03, 0x00, 0x04},
		},
		{
			name: "ReadInputRegistersRequest",
			adu: &modbusApplicationDataUnit{
				header: network.NewHeader([]byte{0x00, 0x01}, []byte{0x00, 0x00}, 0x01),
				pdu:    transport.NewProtocolDataUnit(&data.ReadInputRegistersRequest{Offset: 0, Count: 4}),
			},
			expectedBytes: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x05, 0x01, 0x04, 0x00, 0x00, 0x00, 0x04},
		},
		{
			name: "ReadInputRegistersResponse",
			adu: &modbusApplicationDataUnit{
				header: network.NewHeader([]byte{0x00, 0x01}, []byte{0x00, 0x00}, 0x01),
				pdu:    transport.NewProtocolDataUnit(&data.ReadInputRegistersResponse{Values: []uint16{0x0001, 0x0002, 0x0003, 0x0004}}),
			},
			expectedBytes: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x0A, 0x01, 0x04, 0x08, 0x00, 0x01, 0x00, 0x02, 0x00, 0x03, 0x00, 0x04},
		},
		{
			name: "WriteSingleCoilRequest",
			adu: &modbusApplicationDataUnit{
				header: network.NewHeader([]byte{0x00, 0x01}, []byte{0x00, 0x00}, 0x01),
				pdu:    transport.NewProtocolDataUnit(&data.WriteSingleCoilRequest{Offset: 0, Value: true}),
			},
			expectedBytes: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x05, 0x01, 0x05, 0x00, 0x00, 0xFF, 0x00},
		},
		{
			name: "WriteSingleCoilResponse",
			adu: &modbusApplicationDataUnit{
				header: network.NewHeader([]byte{0x00, 0x01}, []byte{0x00, 0x00}, 0x01),
				pdu:    transport.NewProtocolDataUnit(&data.WriteSingleCoilResponse{Offset: 0, Value: true}),
			},
			expectedBytes: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x05, 0x01, 0x05, 0x00, 0x00, 0xFF, 0x00},
		},
		{
			name: "WriteSingleRegisterRequest",
			adu: &modbusApplicationDataUnit{
				header: network.NewHeader([]byte{0x00, 0x01}, []byte{0x00, 0x00}, 0x01),
				pdu:    transport.NewProtocolDataUnit(&data.WriteSingleRegisterRequest{Offset: 0, Value: 0x0001}),
			},
			expectedBytes: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x05, 0x01, 0x06, 0x00, 0x00, 0x00, 0x01},
		},
		{
			name: "WriteSingleRegisterResponse",
			adu: &modbusApplicationDataUnit{
				header: network.NewHeader([]byte{0x00, 0x01}, []byte{0x00, 0x00}, 0x01),
				pdu:    transport.NewProtocolDataUnit(&data.WriteSingleRegisterResponse{Offset: 0, Value: 0x0001}),
			},
			expectedBytes: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x05, 0x01, 0x06, 0x00, 0x00, 0x00, 0x01},
		},
		{
			name: "WriteMultipleCoilsRequest",
			adu: &modbusApplicationDataUnit{
				header: network.NewHeader([]byte{0x00, 0x01}, []byte{0x00, 0x00}, 0x01),
				pdu: transport.NewProtocolDataUnit(&data.WriteMultipleCoilsRequest{
					Offset: 0,
					Values: []bool{true, false, false, false, false, false, false, true, true, false, false, false, false, false, false, true},
				}),
			},
			expectedBytes: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x08, 0x01, 0x0F, 0x00, 0x00, 0x00, 0x10, 0x02, 0x81, 0x81},
		},
		{
			name: "WriteMultipleCoilsResponse",
			adu: &modbusApplicationDataUnit{
				header: network.NewHeader([]byte{0x00, 0x01}, []byte{0x00, 0x00}, 0x01),
				pdu: transport.NewProtocolDataUnit(&data.WriteMultipleCoilsResponse{
					Offset: 0,
					Count:  16,
				}),
			},
			expectedBytes: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x05, 0x01, 0x0F, 0x00, 0x00, 0x00, 0x10},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedBytes, tt.adu.Bytes())
		})
	}
}
