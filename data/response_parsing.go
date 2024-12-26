package data

import (
	"github.com/rinzlerlabs/gomodbus/common"
)

func ParseModbusResponseOperation(functionCode FunctionCode, bytes []byte, valueCount int) (ModbusOperation, error) {
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
		op, err = NewModbusOperationExceptionFromResponse(functionCode, bytes)
	case ReadDiscreteInputsError:
		op, err = NewModbusOperationExceptionFromResponse(functionCode, bytes)
	case ReadHoldingRegistersError:
		op, err = NewModbusOperationExceptionFromResponse(functionCode, bytes)
	case ReadInputRegistersError:
		op, err = NewModbusOperationExceptionFromResponse(functionCode, bytes)
	case WriteSingleCoilError:
		op, err = NewModbusOperationExceptionFromResponse(functionCode, bytes)
	case WriteSingleRegisterError:
		op, err = NewModbusOperationExceptionFromResponse(functionCode, bytes)
	case WriteMultipleCoilsError:
		op, err = NewModbusOperationExceptionFromResponse(functionCode, bytes)
	case WriteMultipleRegistersError:
		op, err = NewModbusOperationExceptionFromResponse(functionCode, bytes)
	default:
		return nil, common.ErrInvalidFunctionCode
	}
	return op, err
}

func newReadCoilsResponse(b []byte, requestCount int) (*ReadCoilsResponse, error) {
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
		values: values[:requestCount],
	}, nil
}

func newReadDiscreteInputsResponse(b []byte, requestCount int) (*ReadDiscreteInputsResponse, error) {
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
		values: values[:requestCount],
	}, nil
}

func newReadHoldingRegistersResponse(b []byte, requestCount int) (*ReadHoldingRegistersResponse, error) {
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
		values: values[:requestCount],
	}, nil
}

func newReadInputRegistersResponse(b []byte, requestCount int) (*ReadInputRegistersResponse, error) {
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
		values: values[:requestCount],
	}, nil
}

func newWriteSingleCoilResponse(b []byte) (*WriteSingleCoilResponse, error) {
	if len(b) != 4 {
		return nil, common.ErrInvalidPacket
	}
	return &WriteSingleCoilResponse{
		offset: uint16(b[0])<<8 | uint16(b[1]),
		value:  b[2] == 0xFF,
	}, nil
}

func newWriteSingleRegisterResponse(b []byte) (*WriteSingleRegisterResponse, error) {
	if len(b) != 4 {
		return nil, common.ErrInvalidPacket
	}
	return &WriteSingleRegisterResponse{
		offset: uint16(b[0])<<8 | uint16(b[1]),
		value:  uint16(b[2])<<8 | uint16(b[3]),
	}, nil
}

func newWriteMultipleCoilsResponse(b []byte) (*WriteMultipleCoilsResponse, error) {
	if len(b) != 4 {
		return nil, common.ErrInvalidPacket
	}
	return &WriteMultipleCoilsResponse{
		offset: uint16(b[0])<<8 | uint16(b[1]),
		count:  uint16(b[2])<<8 | uint16(b[3]),
	}, nil
}

func newWriteMultipleRegistersResponse(b []byte) (*WriteMultipleRegistersResponse, error) {
	if len(b) != 4 {
		return nil, common.ErrInvalidPacket
	}
	return &WriteMultipleRegistersResponse{
		offset: uint16(b[0])<<8 | uint16(b[1]),
		count:  uint16(b[2])<<8 | uint16(b[3]),
	}, nil
}

func NewModbusOperationExceptionFromResponse(functionCode FunctionCode, b []byte) (*ModbusOperationException, error) {
	if len(b) != 1 {
		return nil, common.ErrInvalidPacket
	}
	return &ModbusOperationException{
		FunctionCode:  functionCode,
		ExceptionCode: ExceptionCode(b[0]),
	}, nil
}
