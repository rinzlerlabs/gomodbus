package data

import (
	"github.com/rinzlerlabs/gomodbus/common"
)

func ParseModbusResponseOperation(functionCode FunctionCode, bytes []byte, valueCount uint16) (ModbusOperation, error) {
	var op ModbusOperation
	var err error
	switch functionCode {
	case ReadCoils:
		op, err = newReadCoilsResponse(bytes, valueCount)
	case ReadDiscreteInputs:
		op, err = newReadDiscreteInputsResponse(bytes, valueCount)
	case ReadHoldingRegisters:
		op, err = newReadHoldingRegistersResponse(bytes, valueCount)
	case ReadInputRegisters:
		op, err = newReadInputRegistersResponse(bytes, valueCount)
	case WriteSingleCoil:
		op, err = newWriteSingleCoilResponse(bytes)
	case WriteSingleRegister:
		op, err = newWriteSingleRegisterResponse(bytes)
	case WriteMultipleCoils:
		op, err = newWriteMultipleCoilsResponse(bytes)
	case WriteMultipleRegisters:
		op, err = newWriteMultipleRegistersResponse(bytes)
	case ReadCoilsError:
		op, err = NewModbusOperationExceptionFromResponse(bytes)
	default:
		return nil, common.ErrInvalidFunctionCode
	}
	return op, err
}

func newReadCoilsResponse(b []byte, requestCount uint16) (*ReadCoilsResponse, error) {
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
	return &ReadCoilsResponse{
		Values: values[:requestCount],
	}, nil
}

func newReadDiscreteInputsResponse(b []byte, requestCount uint16) (*ReadDiscreteInputsResponse, error) {
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
	return &ReadDiscreteInputsResponse{
		Values: values[:requestCount],
	}, nil
}

func newReadHoldingRegistersResponse(b []byte, requestCount uint16) (*ReadHoldingRegistersResponse, error) {
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
	return &ReadHoldingRegistersResponse{
		Values: values[:requestCount],
	}, nil
}

func newReadInputRegistersResponse(b []byte, requestCount uint16) (*ReadInputRegistersResponse, error) {
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
	return &ReadInputRegistersResponse{
		Values: values[:requestCount],
	}, nil
}

func newWriteSingleCoilResponse(b []byte) (*WriteSingleCoilResponse, error) {
	if len(b) != 4 {
		return nil, common.ErrInvalidPacket
	}
	return &WriteSingleCoilResponse{
		Offset: uint16(b[0])<<8 | uint16(b[1]),
		Value:  b[2] == 0xFF,
	}, nil
}

func newWriteSingleRegisterResponse(b []byte) (*WriteSingleRegisterResponse, error) {
	if len(b) != 4 {
		return nil, common.ErrInvalidPacket
	}
	return &WriteSingleRegisterResponse{
		Offset: uint16(b[0])<<8 | uint16(b[1]),
		Value:  uint16(b[2])<<8 | uint16(b[3]),
	}, nil
}

func newWriteMultipleCoilsResponse(b []byte) (*WriteMultipleCoilsResponse, error) {
	if len(b) != 4 {
		return nil, common.ErrInvalidPacket
	}
	return &WriteMultipleCoilsResponse{
		Offset: uint16(b[0])<<8 | uint16(b[1]),
		Count:  uint16(b[2])<<8 | uint16(b[3]),
	}, nil
}

func newWriteMultipleRegistersResponse(b []byte) (*WriteMultipleRegistersResponse, error) {
	if len(b) != 4 {
		return nil, common.ErrInvalidPacket
	}
	return &WriteMultipleRegistersResponse{
		Offset: uint16(b[0])<<8 | uint16(b[1]),
		Count:  uint16(b[2])<<8 | uint16(b[3]),
	}, nil
}

func NewModbusOperationExceptionFromResponse(b []byte) (*ModbusOperationException, error) {
	if len(b) != 1 {
		return nil, common.ErrInvalidPacket
	}
	return &ModbusOperationException{
		ExceptionCode: ExceptionCode(b[0]),
	}, nil
}
