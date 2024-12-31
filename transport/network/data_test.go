package network

import (
	"testing"

	"github.com/rinzlerlabs/gomodbus/data"
	"github.com/rinzlerlabs/gomodbus/transport"
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
				header: NewHeader([]byte{0x00, 0x01}, []byte{0x00, 0x00}, 0x01),
				pdu:    transport.NewProtocolDataUnit(data.NewReadCoilsRequest(0, 16)),
			},
			expectedBytes: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x01, 0x00, 0x00, 0x00, 0x10},
		},
		{
			name: "ReadCoilsResponse",
			adu: &modbusApplicationDataUnit{
				header: NewHeader([]byte{0x00, 0x01}, []byte{0x00, 0x00}, 0x01),
				pdu:    transport.NewProtocolDataUnit(data.NewReadCoilsResponse([]bool{true, false, false, false, false, false, false, true, true, false, false, false, false, false, false, true})),
			},
			expectedBytes: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x05, 0x01, 0x01, 0x02, 0x81, 0x81},
		},
		{
			name: "ReadDiscreteInputsRequest",
			adu: &modbusApplicationDataUnit{
				header: NewHeader([]byte{0x00, 0x01}, []byte{0x00, 0x00}, 0x01),
				pdu:    transport.NewProtocolDataUnit(data.NewReadDiscreteInputsRequest(0, 16)),
			},
			expectedBytes: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x02, 0x00, 0x00, 0x00, 0x10},
		},
		{
			name: "ReadDiscreteInputsResponse",
			adu: &modbusApplicationDataUnit{
				header: NewHeader([]byte{0x00, 0x01}, []byte{0x00, 0x00}, 0x01),
				pdu:    transport.NewProtocolDataUnit(data.NewReadDiscreteInputsResponse([]bool{true, false, false, false, false, false, false, true, true, false, false, false, false, false, false, true})),
			},
			expectedBytes: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x05, 0x01, 0x02, 0x02, 0x81, 0x81},
		},
		{
			name: "ReadHoldingRegistersRequest",
			adu: &modbusApplicationDataUnit{
				header: NewHeader([]byte{0x00, 0x01}, []byte{0x00, 0x00}, 0x01),
				pdu:    transport.NewProtocolDataUnit(data.NewReadHoldingRegistersRequest(0, 4)),
			},
			expectedBytes: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x03, 0x00, 0x00, 0x00, 0x04},
		},
		{
			name: "ReadHoldingRegistersResponse",
			adu: &modbusApplicationDataUnit{
				header: NewHeader([]byte{0x00, 0x01}, []byte{0x00, 0x00}, 0x01),
				pdu:    transport.NewProtocolDataUnit(data.NewReadHoldingRegistersResponse([]uint16{0x0001, 0x0002, 0x0003, 0x0004})),
			},
			expectedBytes: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x0B, 0x01, 0x03, 0x08, 0x00, 0x01, 0x00, 0x02, 0x00, 0x03, 0x00, 0x04},
		},
		{
			name: "ReadInputRegistersRequest",
			adu: &modbusApplicationDataUnit{
				header: NewHeader([]byte{0x00, 0x01}, []byte{0x00, 0x00}, 0x01),
				pdu:    transport.NewProtocolDataUnit(data.NewReadInputRegistersRequest(0, 4)),
			},
			expectedBytes: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x04, 0x00, 0x00, 0x00, 0x04},
		},
		{
			name: "ReadInputRegistersResponse",
			adu: &modbusApplicationDataUnit{
				header: NewHeader([]byte{0x00, 0x01}, []byte{0x00, 0x00}, 0x01),
				pdu:    transport.NewProtocolDataUnit(data.NewReadInputRegistersResponse([]uint16{0x0001, 0x0002, 0x0003, 0x0004})),
			},
			expectedBytes: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x0B, 0x01, 0x04, 0x08, 0x00, 0x01, 0x00, 0x02, 0x00, 0x03, 0x00, 0x04},
		},
		{
			name: "WriteSingleCoilRequest",
			adu: &modbusApplicationDataUnit{
				header: NewHeader([]byte{0x00, 0x01}, []byte{0x00, 0x00}, 0x01),
				pdu:    transport.NewProtocolDataUnit(data.NewWriteSingleCoilRequest(0, true)),
			},
			expectedBytes: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x05, 0x00, 0x00, 0xFF, 0x00},
		},
		{
			name: "WriteSingleCoilResponse",
			adu: &modbusApplicationDataUnit{
				header: NewHeader([]byte{0x00, 0x01}, []byte{0x00, 0x00}, 0x01),
				pdu:    transport.NewProtocolDataUnit(data.NewWriteSingleCoilResponse(0, true)),
			},
			expectedBytes: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x05, 0x00, 0x00, 0xFF, 0x00},
		},
		{
			name: "WriteSingleRegisterRequest",
			adu: &modbusApplicationDataUnit{
				header: NewHeader([]byte{0x00, 0x01}, []byte{0x00, 0x00}, 0x01),
				pdu:    transport.NewProtocolDataUnit(data.NewWriteSingleRegisterRequest(0, 0x0001)),
			},
			expectedBytes: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x06, 0x00, 0x00, 0x00, 0x01},
		},
		{
			name: "WriteSingleRegisterResponse",
			adu: &modbusApplicationDataUnit{
				header: NewHeader([]byte{0x00, 0x01}, []byte{0x00, 0x00}, 0x01),
				pdu:    transport.NewProtocolDataUnit(data.NewWriteSingleRegisterResponse(0, 0x0001)),
			},
			expectedBytes: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x06, 0x00, 0x00, 0x00, 0x01},
		},
		{
			name: "WriteMultipleCoilsRequest",
			adu: &modbusApplicationDataUnit{
				header: NewHeader([]byte{0x00, 0x01}, []byte{0x00, 0x00}, 0x01),
				pdu: transport.NewProtocolDataUnit(data.NewWriteMultipleCoilsRequest(
					0,
					[]bool{true, false, false, false, false, false, false, true, true, false, false, false, false, false, false, true},
				)),
			},
			expectedBytes: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x09, 0x01, 0x0F, 0x00, 0x00, 0x00, 0x10, 0x02, 0x81, 0x81},
		},
		{
			name: "WriteMultipleCoilsResponse",
			adu: &modbusApplicationDataUnit{
				header: NewHeader([]byte{0x00, 0x01}, []byte{0x00, 0x00}, 0x01),
				pdu:    transport.NewProtocolDataUnit(data.NewWriteMultipleCoilsResponse(0, 16)),
			},
			expectedBytes: []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x0F, 0x00, 0x00, 0x00, 0x10},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedBytes, tt.adu.Bytes())
		})
	}
}
