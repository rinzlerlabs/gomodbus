package data

import "github.com/rinzlerlabs/gomodbus/common"

func ParseModbusRequestOperation(functionCode FunctionCode, bytes []byte) (ModbusOperation, error) {
	var op ModbusOperation
	var err error
	switch functionCode {
	case ReadCoils:
		op, err = newReadCoilsRequest(bytes)
	case ReadDiscreteInputs:
		op, err = newReadDiscreteInputsRequest(bytes)
	case ReadHoldingRegisters:
		op, err = newReadHoldingRegistersRequest(bytes)
	case ReadInputRegisters:
		op, err = newReadInputRegistersRequest(bytes)
	case WriteSingleCoil:
		op, err = newWriteSingleCoilRequest(bytes)
	case WriteSingleRegister:
		op, err = newWriteSingleRegisterRequest(bytes)
	case WriteMultipleCoils:
		op, err = newWriteMultipleCoilsRequest(bytes)
	case WriteMultipleRegisters:
		op, err = newWriteMultipleRegistersRequest(bytes)
	default:
		return nil, common.ErrInvalidFunctionCode
	}
	return op, err
}

func newReadCoilsRequest(bytes []byte) (*ReadCoilsRequest, error) {
	if len(bytes) != 4 {
		return nil, common.ErrInvalidPacket
	}
	return &ReadCoilsRequest{
		Offset: uint16(bytes[0])<<8 | uint16(bytes[1]),
		Count:  uint16(bytes[2])<<8 | uint16(bytes[3]),
	}, nil
}

func newReadDiscreteInputsRequest(bytes []byte) (*ReadDiscreteInputsRequest, error) {
	if len(bytes) != 4 {
		return nil, common.ErrInvalidPacket
	}
	return &ReadDiscreteInputsRequest{
		Offset: uint16(bytes[0])<<8 | uint16(bytes[1]),
		Count:  uint16(bytes[2])<<8 | uint16(bytes[3]),
	}, nil
}

func newReadHoldingRegistersRequest(bytes []byte) (*ReadHoldingRegistersRequest, error) {
	if len(bytes) != 4 {
		return nil, common.ErrInvalidPacket
	}
	return &ReadHoldingRegistersRequest{
		Offset: uint16(bytes[0])<<8 | uint16(bytes[1]),
		Count:  uint16(bytes[2])<<8 | uint16(bytes[3]),
	}, nil
}

func newReadInputRegistersRequest(bytes []byte) (*ReadInputRegistersRequest, error) {
	if len(bytes) != 4 {
		return nil, common.ErrInvalidPacket
	}
	return &ReadInputRegistersRequest{
		Offset: uint16(bytes[0])<<8 | uint16(bytes[1]),
		Count:  uint16(bytes[2])<<8 | uint16(bytes[3]),
	}, nil
}

func newWriteSingleCoilRequest(bytes []byte) (*WriteSingleCoilRequest, error) {
	if len(bytes) != 4 {
		return nil, common.ErrInvalidPacket
	}
	return &WriteSingleCoilRequest{
		Offset: uint16(bytes[0])<<8 | uint16(bytes[1]),
		Value:  bytes[2] == 0xFF && bytes[3] == 0x00,
	}, nil
}

func newWriteSingleRegisterRequest(bytes []byte) (*WriteSingleRegisterRequest, error) {
	if len(bytes) != 4 {
		return nil, common.ErrInvalidPacket
	}
	return &WriteSingleRegisterRequest{
		Offset: uint16(bytes[0])<<8 | uint16(bytes[1]),
		Value:  uint16(bytes[2])<<8 | uint16(bytes[3]),
	}, nil
}

func newWriteMultipleCoilsRequest(bytes []byte) (*WriteMultipleCoilsRequest, error) {
	if len(bytes) < 5 {
		return nil, common.ErrInvalidPacket
	}
	offset := uint16(bytes[0])<<8 | uint16(bytes[1])
	coilCount := uint16(bytes[2])<<8 | uint16(bytes[3])
	byteCount := uint16(bytes[4])
	if uint16(len(bytes)) != 2+2+byteCount+1 {
		return nil, common.ErrInvalidPacket
	}
	if byteCount*8 < coilCount {
		return nil, common.ErrInvalidPacket
	}
	values := make([]bool, coilCount)
	for i := uint16(0); i < coilCount; i++ {
		values[i] = bytes[5+i/8]&(1<<uint(i%8)) != 0
	}
	return &WriteMultipleCoilsRequest{
		Offset: offset,
		Values: values,
	}, nil
}

func newWriteMultipleRegistersRequest(bytes []byte) (*WriteMultipleRegistersRequest, error) {
	if len(bytes) < 5 {
		return nil, common.ErrInvalidPacket
	}
	offset := uint16(bytes[0])<<8 | uint16(bytes[1])
	registerCount := uint16(bytes[2])<<8 | uint16(bytes[3])
	byteCount := uint16(bytes[4])
	if uint16(len(bytes)) != 2+2+byteCount+1 {
		return nil, common.ErrInvalidPacket
	}
	if byteCount != registerCount*2 {
		return nil, common.ErrInvalidPacket
	}
	values := make([]uint16, registerCount)
	for i := uint16(0); i < registerCount; i++ {
		values[i] = uint16(bytes[5+i*2])<<8 | uint16(bytes[6+i*2])
	}
	return &WriteMultipleRegistersRequest{
		Offset: offset,
		Values: values,
	}, nil
}