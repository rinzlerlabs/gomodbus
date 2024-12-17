package gomodbus

import (
	"testing"

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
		{name: "InvalidRequest_MissingTrailers", data: ":0401000A000DE4", eError: ErrInvalidPacket},
		{name: "InvalidRequest_InvalidStart", data: "0401000A000DE4\r\n", eError: ErrInvalidPacket},
		{name: "InvalidRequest_InvalidChecksum", data: ":0401000A000DE5\r\n", eError: ErrInvalidChecksum},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adu, err := NewASCIIApplicationDataUnitFromRequest(tt.data)
			assert.Equal(t, tt.eError, err)
			if tt.eError == nil {
				assert.Equal(t, tt.eAddress, adu.Address())
				assert.Equal(t, tt.eFunction, adu.PDU().Function)
				assert.Equal(t, tt.eData, adu.PDU().Data)
				assert.Equal(t, tt.eChecksum, adu.Checksum())
			}
		})
	}
}

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
			adu, err := NewRTUApplicationDataUnitFromRequest(tt.data)
			assert.Equal(t, tt.eError, err)
			if tt.eError == nil {
				assert.Equal(t, tt.eAddress, adu.Address())
				assert.Equal(t, tt.eFunction, adu.PDU().Function)
				assert.Equal(t, tt.eData, adu.PDU().Data)
				assert.Equal(t, tt.eChecksum, adu.Checksum())
			}
		})
	}
}
