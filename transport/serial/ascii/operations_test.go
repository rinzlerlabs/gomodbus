package ascii

import (
	"testing"

	"github.com/rinzlerlabs/gomodbus/common"
	"github.com/rinzlerlabs/gomodbus/data"
	"github.com/stretchr/testify/assert"
)

func TestNewApplicationDataUnitFromASCIIRequest(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		eAddress  uint16
		eFunction byte
		eData     []byte
		eChecksum []byte
		eError    error
	}{
		{name: "ValidRequest", data: ":0401000A000DE4\r\n", eAddress: 0x04, eFunction: 0x01, eData: []byte{0x00, 0x0A, 0x00, 0x0D}, eChecksum: []byte{0xE4}, eError: nil},
		{name: "InvalidRequest_MissingTrailers", data: ":0401000A000DE4", eError: common.ErrInvalidPacket},
		{name: "InvalidRequest_InvalidStart", data: "0401000A000DE4\r\n", eError: common.ErrInvalidPacket},
		{name: "InvalidRequest_InvalidChecksum", data: ":0401000A000DE5\r\n", eError: common.ErrInvalidChecksum},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adu, err := NewApplicationDataUnitFromWire(tt.data)
			assert.Equal(t, tt.eError, err)
			if tt.eError == nil {
				assert.Equal(t, tt.eAddress, adu.Address())
				assert.Equal(t, data.FunctionCode(tt.eFunction), adu.PDU().Function)
				assert.Equal(t, tt.eData, adu.PDU().Data)
				assert.Equal(t, tt.eChecksum, adu.Checksum())
			}
		})
	}
}
