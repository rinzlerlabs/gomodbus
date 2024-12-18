package client

import (
	"github.com/rinzlerlabs/gomodbus/common"
	"github.com/rinzlerlabs/gomodbus/data"
	"github.com/rinzlerlabs/gomodbus/transport/serial/ascii"
	"github.com/rinzlerlabs/gomodbus/transport/serial/rtu"
)

func newRTUApplicationDataUnitFromModbusRequest(address uint16, functionCode data.FunctionCode, req data.ModbusRequest) data.ModbusFrame {
	pdu := &data.ProtocolDataUnit{
		Function: functionCode,
		Data:     req.Bytes(),
	}
	return rtu.NewApplicationDataUnitFromRequest(address, pdu)
}

func newASCIIApplicationDataUnitFromModbusRequest(address uint16, functionCode data.FunctionCode, req data.ModbusRequest) data.ModbusFrame {
	pdu := &data.ProtocolDataUnit{
		Function: functionCode,
		Data:     req.Bytes(),
	}
	return ascii.NewApplicationDataUnitFromRequest(address, pdu)
}

func newReadCoilsResponse(responseADU data.ApplicationDataUnit, requestCount uint16) (*data.ReadCoilsResponse, error) {
	b := responseADU.PDU().Data
	if len(b) < 1 {
		return nil, common.ErrInvalidPacket
	}
	byteCount := b[0]
	if len(b) != 1+int(byteCount) {
		return nil, common.ErrInvalidPacket
	}
	values := make([]bool, 8*byteCount)
	for i := 0; i < 8*int(byteCount); i++ {
		values[i] = b[1+i/8]&(1<<uint(i%8)) != 0
	}
	return &data.ReadCoilsResponse{
		Values: values[:requestCount],
	}, nil
}

func newReadDiscreteInputsResponse(responseADU data.ApplicationDataUnit, requestCount uint16) (*data.ReadDiscreteInputsResponse, error) {
	b := responseADU.PDU().Data
	if len(b) < 1 {
		return nil, common.ErrInvalidPacket
	}
	byteCount := b[0]
	if len(b) != 1+int(byteCount) {
		return nil, common.ErrInvalidPacket
	}
	values := make([]bool, 8*byteCount)
	for i := 0; i < 8*int(byteCount); i++ {
		values[i] = b[1+i/8]&(1<<uint(i%8)) != 0
	}
	return &data.ReadDiscreteInputsResponse{
		Values: values[:requestCount],
	}, nil
}

func newReadHoldingRegistersResponse(responseADU data.ApplicationDataUnit, requestCount uint16) (*data.ReadHoldingRegistersResponse, error) {
	b := responseADU.PDU().Data
	if len(b) < 1 {
		return nil, common.ErrInvalidPacket
	}
	byteCount := b[0]
	if len(b) != 1+int(byteCount) {
		return nil, common.ErrInvalidPacket
	}
	values := make([]uint16, byteCount/2)
	for i := 0; i < len(values); i++ {
		values[i] = uint16(b[1+2*i])<<8 | uint16(b[2+2*i])
	}
	return &data.ReadHoldingRegistersResponse{
		Values: values[:requestCount],
	}, nil
}

func newReadInputRegistersResponse(responseADU data.ApplicationDataUnit, requestCount uint16) (*data.ReadInputRegistersResponse, error) {
	b := responseADU.PDU().Data
	if len(b) < 1 {
		return nil, common.ErrInvalidPacket
	}
	byteCount := b[0]
	if len(b) != 1+int(byteCount) {
		return nil, common.ErrInvalidPacket
	}
	values := make([]uint16, byteCount/2)
	for i := 0; i < len(values); i++ {
		values[i] = uint16(b[1+2*i])<<8 | uint16(b[2+2*i])
	}
	return &data.ReadInputRegistersResponse{
		Values: values[:requestCount],
	}, nil
}

func newWriteSingleCoilResponse(responseADU data.ApplicationDataUnit) (*data.WriteSingleCoilResponse, error) {
	b := responseADU.PDU().Data
	if len(b) != 4 {
		return nil, common.ErrInvalidPacket
	}
	return &data.WriteSingleCoilResponse{
		Offset: uint16(b[0])<<8 | uint16(b[1]),
		Value:  b[2] == 0xFF,
	}, nil
}

func newWriteSingleRegisterResponse(responseADU data.ApplicationDataUnit) (*data.WriteSingleRegisterResponse, error) {
	b := responseADU.PDU().Data
	if len(b) != 4 {
		return nil, common.ErrInvalidPacket
	}
	return &data.WriteSingleRegisterResponse{
		Offset: uint16(b[0])<<8 | uint16(b[1]),
		Value:  uint16(b[2])<<8 | uint16(b[3]),
	}, nil
}

func newWriteMultipleCoilsResponse(responseADU data.ApplicationDataUnit) (*data.WriteMultipleCoilsResponse, error) {
	b := responseADU.PDU().Data
	if len(b) != 4 {
		return nil, common.ErrInvalidPacket
	}
	return &data.WriteMultipleCoilsResponse{
		Offset: uint16(b[0])<<8 | uint16(b[1]),
		Count:  uint16(b[2])<<8 | uint16(b[3]),
	}, nil
}

func newWriteMultipleRegistersResponse(responseADU data.ApplicationDataUnit) (*data.WriteMultipleRegistersResponse, error) {
	b := responseADU.PDU().Data
	if len(b) != 4 {
		return nil, common.ErrInvalidPacket
	}
	return &data.WriteMultipleRegistersResponse{
		Offset: uint16(b[0])<<8 | uint16(b[1]),
		Count:  uint16(b[2])<<8 | uint16(b[3]),
	}, nil
}
