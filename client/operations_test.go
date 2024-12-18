package client

import (
	"testing"

	"github.com/rinzlerlabs/gomodbus/data"
	"github.com/rinzlerlabs/gomodbus/transport/serial/rtu"
	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	response := &data.ReadCoilsResponse{
		Values: []bool{true, false, true, false, true, false, true, false},
	}
	pdu := &data.ProtocolDataUnit{
		Function: 0x01,
		Data:     response.Bytes(),
	}
	adu := rtu.NewApplicationDataUnitFromRequest(0x01, pdu)

	res, err := rtu.NewApplicationDataUnitFromWire(adu.Bytes())
	assert.NoError(t, err)
	response, err = newReadCoilsResponse(res, 8)
	assert.NoError(t, err)
	assert.Equal(t, []bool{true, false, true, false, true, false, true, false}, response.Values)
}
