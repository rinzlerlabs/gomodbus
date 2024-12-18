package rtu

import (
	"testing"

	"github.com/rinzlerlabs/gomodbus/data"
	"github.com/stretchr/testify/assert"
)

func TestNewApplicationDataUnitFromRTURequest(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		eAddress  uint16
		eFunction byte
		eData     []byte
		eChecksum []byte
		eError    error
	}{
		{name: "ValidRequest", data: []byte{0x04, 0x01, 0x00, 0x0A, 0x00, 0x0D, 0xDD, 0x98}, eAddress: 0x04, eFunction: 0x01, eData: []byte{0x00, 0x0A, 0x00, 0x0D}, eChecksum: []byte{0xDD, 0x98}, eError: nil},
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
