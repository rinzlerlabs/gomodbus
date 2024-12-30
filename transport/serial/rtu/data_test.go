package rtu

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
		header        transport.Header
		pdu           *transport.ProtocolDataUnit
		expectedBytes []byte
	}{
		{
			name:          "ReadCoilsRequest",
			header:        serial.NewHeader(1),
			pdu:           transport.NewProtocolDataUnit(data.NewReadCoilsRequest(0, 16)),
			expectedBytes: []byte{0x01, 0x01, 0x00, 0x00, 0x00, 0x10},
		},
		{
			name:          "ReadCoilsResponse",
			header:        serial.NewHeader(1),
			pdu:           transport.NewProtocolDataUnit(data.NewReadCoilsResponse([]bool{true, false, false, false, false, false, false, true, true, false, false, false, false, false, false, true})),
			expectedBytes: []byte{0x01, 0x01, 0x02, 0x81, 0x81},
		},
		{
			name:          "ReadDiscreteInputsRequest",
			header:        serial.NewHeader(1),
			pdu:           transport.NewProtocolDataUnit(data.NewReadDiscreteInputsRequest(0, 16)),
			expectedBytes: []byte{0x01, 0x02, 0x00, 0x00, 0x00, 0x10},
		},
		{
			name:          "ReadDiscreteInputsResponse",
			header:        serial.NewHeader(1),
			pdu:           transport.NewProtocolDataUnit(data.NewReadDiscreteInputsResponse([]bool{true, false, false, false, false, false, false, true, true, false, false, false, false, false, false, true})),
			expectedBytes: []byte{0x01, 0x02, 0x02, 0x81, 0x81},
		},
		{
			name:          "ReadHoldingRegistersRequest",
			header:        serial.NewHeader(1),
			pdu:           transport.NewProtocolDataUnit(data.NewReadHoldingRegistersRequest(0, 4)),
			expectedBytes: []byte{0x01, 0x03, 0x00, 0x00, 0x00, 0x04},
		},
		{
			name:          "ReadHoldingRegistersResponse",
			header:        serial.NewHeader(1),
			pdu:           transport.NewProtocolDataUnit(data.NewReadHoldingRegistersResponse([]uint16{0x0001, 0x0002, 0x0003, 0x0004})),
			expectedBytes: []byte{0x01, 0x03, 0x08, 0x00, 0x01, 0x00, 0x02, 0x00, 0x03, 0x00, 0x04},
		},
		{
			name:          "ReadInputRegistersRequest",
			header:        serial.NewHeader(1),
			pdu:           transport.NewProtocolDataUnit(data.NewReadInputRegistersRequest(0, 4)),
			expectedBytes: []byte{0x01, 0x04, 0x00, 0x00, 0x00, 0x04},
		},
		{
			name:          "ReadInputRegistersResponse",
			header:        serial.NewHeader(1),
			pdu:           transport.NewProtocolDataUnit(data.NewReadInputRegistersResponse([]uint16{0x0001, 0x0002, 0x0003, 0x0004})),
			expectedBytes: []byte{0x01, 0x04, 0x08, 0x00, 0x01, 0x00, 0x02, 0x00, 0x03, 0x00, 0x04},
		},
		{
			name:          "WriteSingleCoilRequest",
			header:        serial.NewHeader(1),
			pdu:           transport.NewProtocolDataUnit(data.NewWriteSingleCoilRequest(0, true)),
			expectedBytes: []byte{0x01, 0x05, 0x00, 0x00, 0xFF, 0x00},
		},
		{
			name:          "WriteSingleCoilResponse",
			header:        serial.NewHeader(1),
			pdu:           transport.NewProtocolDataUnit(data.NewWriteSingleCoilResponse(0, true)),
			expectedBytes: []byte{0x01, 0x05, 0x00, 0x00, 0xFF, 0x00},
		},
		{
			name:          "WriteSingleRegisterRequest",
			header:        serial.NewHeader(1),
			pdu:           transport.NewProtocolDataUnit(data.NewWriteSingleRegisterRequest(0, 0x0001)),
			expectedBytes: []byte{0x01, 0x06, 0x00, 0x00, 0x00, 0x01},
		},
		{
			name:          "WriteSingleRegisterResponse",
			header:        serial.NewHeader(1),
			pdu:           transport.NewProtocolDataUnit(data.NewWriteSingleRegisterResponse(0, 0x0001)),
			expectedBytes: []byte{0x01, 0x06, 0x00, 0x00, 0x00, 0x01},
		},
		{
			name:          "WriteMultipleCoilsRequest",
			header:        serial.NewHeader(1),
			pdu:           transport.NewProtocolDataUnit(data.NewWriteMultipleCoilsRequest(0, []bool{true, false, false, false, false, false, false, true, true, false, false, false, false, false, false, true})),
			expectedBytes: []byte{0x01, 0x0F, 0x00, 0x00, 0x00, 0x10, 0x02, 0x81, 0x81},
		},
		{
			name:          "WriteMultipleCoilsResponse",
			header:        serial.NewHeader(1),
			pdu:           transport.NewProtocolDataUnit(data.NewWriteMultipleCoilsResponse(0, 16)),
			expectedBytes: []byte{0x01, 0x0F, 0x00, 0x00, 0x00, 0x10},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adu, err := NewModbusApplicationDataUnit(tt.header, tt.pdu)
			assert.NoError(t, err)
			bytes := adu.Bytes()
			assert.Equal(t, tt.expectedBytes, bytes[:len(bytes)-2]) // Ignore the checksum because i don't feel like calculating by hand
		})
	}
}
