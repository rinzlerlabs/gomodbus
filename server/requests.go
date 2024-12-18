package server

import (
	"github.com/rinzlerlabs/gomodbus/common"
	"github.com/rinzlerlabs/gomodbus/data"
)

func newReadCoilsRequest(adu data.ApplicationDataUnit) (*data.ReadCoilsRequest, error) {
	pdu := adu.PDU()
	if len(pdu.Data) != 4 {
		return nil, common.ErrInvalidPacket
	}
	return &data.ReadCoilsRequest{
		Offset: uint16(pdu.Data[0])<<8 | uint16(pdu.Data[1]),
		Count:  uint16(pdu.Data[2])<<8 | uint16(pdu.Data[3]),
	}, nil
}

func newReadDiscreteInputsRequest(adu data.ApplicationDataUnit) (*data.ReadDiscreteInputsRequest, error) {
	pdu := adu.PDU()
	if len(pdu.Data) != 4 {
		return nil, common.ErrInvalidPacket
	}
	return &data.ReadDiscreteInputsRequest{
		Offset: uint16(pdu.Data[0])<<8 | uint16(pdu.Data[1]),
		Count:  uint16(pdu.Data[2])<<8 | uint16(pdu.Data[3]),
	}, nil
}

func newReadHoldingRegistersRequest(adu data.ApplicationDataUnit) (*data.ReadHoldingRegistersRequest, error) {
	pdu := adu.PDU()
	if len(pdu.Data) != 4 {
		return nil, common.ErrInvalidPacket
	}
	return &data.ReadHoldingRegistersRequest{
		Offset: uint16(pdu.Data[0])<<8 | uint16(pdu.Data[1]),
		Count:  uint16(pdu.Data[2])<<8 | uint16(pdu.Data[3]),
	}, nil
}

func newReadInputRegistersRequest(adu data.ApplicationDataUnit) (*data.ReadInputRegistersRequest, error) {
	pdu := adu.PDU()
	if len(pdu.Data) != 4 {
		return nil, common.ErrInvalidPacket
	}
	return &data.ReadInputRegistersRequest{
		Offset: uint16(pdu.Data[0])<<8 | uint16(pdu.Data[1]),
		Count:  uint16(pdu.Data[2])<<8 | uint16(pdu.Data[3]),
	}, nil
}

func newWriteSingleCoilRequest(adu data.ApplicationDataUnit) (*data.WriteSingleCoilRequest, error) {
	pdu := adu.PDU()
	if len(pdu.Data) != 4 {
		return nil, common.ErrInvalidPacket
	}
	return &data.WriteSingleCoilRequest{
		Offset: uint16(pdu.Data[0])<<8 | uint16(pdu.Data[1]),
		Value:  pdu.Data[2] == 0xFF && pdu.Data[3] == 0x00,
	}, nil
}

func newWriteSingleRegisterRequest(adu data.ApplicationDataUnit) (*data.WriteSingleRegisterRequest, error) {
	pdu := adu.PDU()
	if len(pdu.Data) != 4 {
		return nil, common.ErrInvalidPacket
	}
	return &data.WriteSingleRegisterRequest{
		Offset: uint16(pdu.Data[0])<<8 | uint16(pdu.Data[1]),
		Value:  uint16(pdu.Data[2])<<8 | uint16(pdu.Data[3]),
	}, nil
}

func newWriteMultipleCoilsRequest(adu data.ApplicationDataUnit) (*data.WriteMultipleCoilsRequest, error) {
	pdu := adu.PDU()
	if len(pdu.Data) < 5 {
		return nil, common.ErrInvalidPacket
	}
	offset := uint16(pdu.Data[0])<<8 | uint16(pdu.Data[1])
	coilCount := uint16(pdu.Data[2])<<8 | uint16(pdu.Data[3])
	byteCount := uint16(pdu.Data[4])
	if uint16(len(pdu.Data)) != 2+2+byteCount+1 {
		return nil, common.ErrInvalidPacket
	}
	if byteCount*8 < coilCount {
		return nil, common.ErrInvalidPacket
	}
	values := make([]bool, coilCount)
	for i := uint16(0); i < coilCount; i++ {
		values[i] = pdu.Data[5+i/8]&(1<<uint(i%8)) != 0
	}
	return &data.WriteMultipleCoilsRequest{
		Offset: offset,
		Values: values,
	}, nil
}

func newWriteMultipleRegistersRequest(adu data.ApplicationDataUnit) (*data.WriteMultipleRegistersRequest, error) {
	pdu := adu.PDU()
	if len(pdu.Data) < 5 {
		return nil, common.ErrInvalidPacket
	}
	offset := uint16(pdu.Data[0])<<8 | uint16(pdu.Data[1])
	registerCount := uint16(pdu.Data[2])<<8 | uint16(pdu.Data[3])
	byteCount := uint16(pdu.Data[4])
	if uint16(len(pdu.Data)) != 2+2+byteCount+1 {
		return nil, common.ErrInvalidPacket
	}
	if byteCount != registerCount*2 {
		return nil, common.ErrInvalidPacket
	}
	values := make([]uint16, registerCount)
	for i := uint16(0); i < registerCount; i++ {
		values[i] = uint16(pdu.Data[5+i*2])<<8 | uint16(pdu.Data[6+i*2])
	}
	return &data.WriteMultipleRegistersRequest{
		Offset: offset,
		Values: values,
	}, nil
}
