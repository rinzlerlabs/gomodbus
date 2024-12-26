package data

import (
	"github.com/rinzlerlabs/gomodbus/common"
	"go.uber.org/zap/zapcore"
)

type FunctionCode byte
type ExceptionCode byte

const (
	ReadCoils                   FunctionCode = 0x01
	ReadDiscreteInputs          FunctionCode = 0x02
	ReadHoldingRegisters        FunctionCode = 0x03
	ReadInputRegisters          FunctionCode = 0x04
	WriteSingleCoil             FunctionCode = 0x05
	WriteSingleRegister         FunctionCode = 0x06
	WriteMultipleCoils          FunctionCode = 0x0F
	WriteMultipleRegisters      FunctionCode = 0x10
	ReadCoilsError              FunctionCode = 0x81
	ReadDiscreteInputsError     FunctionCode = 0x82
	ReadHoldingRegistersError   FunctionCode = 0x83
	ReadInputRegistersError     FunctionCode = 0x84
	WriteSingleCoilError        FunctionCode = 0x85
	WriteSingleRegisterError    FunctionCode = 0x86
	WriteMultipleCoilsError     FunctionCode = 0x8F
	WriteMultipleRegistersError FunctionCode = 0x90

	IllegalFunction                    ExceptionCode = 0x01
	IllegalDataAddress                 ExceptionCode = 0x02
	IllegalDataValue                   ExceptionCode = 0x03
	ServerDeviceFailure                ExceptionCode = 0x04
	Acknowledge                        ExceptionCode = 0x05
	ServerDeviceBusy                   ExceptionCode = 0x06
	MemoryParityError                  ExceptionCode = 0x08
	GatewayPathUnavailable             ExceptionCode = 0x0A
	GatewayTargetDeviceFailedToRespond ExceptionCode = 0x0B
)

func (f FunctionCode) String() string {
	switch f {
	case ReadCoils:
		return "ReadCoils"
	case ReadDiscreteInputs:
		return "ReadDiscreteInputs"
	case ReadHoldingRegisters:
		return "ReadHoldingRegisters"
	case ReadInputRegisters:
		return "ReadInputRegisters"
	case WriteSingleCoil:
		return "WriteSingleCoil"
	case WriteSingleRegister:
		return "WriteSingleRegister"
	case WriteMultipleCoils:
		return "WriteMultipleCoils"
	case WriteMultipleRegisters:
		return "WriteMultipleRegisters"
	default:
		return "Unknown"
	}
}

func (f FunctionCode) IsException() bool {
	return f >= 0x80 && f <= 0x91
}

func (f ExceptionCode) String() string {
	switch f {
	case IllegalFunction:
		return "IllegalFunction"
	case IllegalDataAddress:
		return "IllegalDataAddress"
	case IllegalDataValue:
		return "IllegalDataValue"
	case ServerDeviceFailure:
		return "SlaveDeviceFailure"
	case Acknowledge:
		return "Acknowledge"
	case ServerDeviceBusy:
		return "SlaveDeviceBusy"
	case MemoryParityError:
		return "MemoryParityError"
	case GatewayPathUnavailable:
		return "GatewayPathUnavailable"
	case GatewayTargetDeviceFailedToRespond:
		return "GatewayTargetDeviceFailedToRespond"
	default:
		return "UnknownException"
	}
}

type ModbusOperation interface {
	zapcore.ObjectMarshaler
}

type CountableOperation interface {
	ModbusOperation
	Count() int
}

func getReturnByteCount(values []bool) byte {
	if len(values)%8 == 0 {
		return byte(len(values) / 8)
	}
	return byte(len(values)/8 + 1)
}

func NewModbusOperationException(requestFunction FunctionCode, code ExceptionCode) *ModbusOperationException {
	return &ModbusOperationException{
		FunctionCode:  requestFunction + 0x80,
		ExceptionCode: code,
	}
}

type ModbusOperationException struct {
	FunctionCode  FunctionCode
	ExceptionCode ExceptionCode
}

func (e ModbusOperationException) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("ExceptionCode", e.ExceptionCode.String())
	return nil
}

func (e *ModbusOperationException) Error() error {
	switch e.ExceptionCode {
	case IllegalFunction:
		return common.ErrIllegalFunction
	case IllegalDataAddress:
		return common.ErrIllegalDataAddress
	case IllegalDataValue:
		return common.ErrIllegalDataValue
	case ServerDeviceFailure:
		return common.ErrServerDeviceFailure
	case Acknowledge:
		return common.ErrAcknowledge
	case ServerDeviceBusy:
		return common.ErrServerDeviceBusy
	case MemoryParityError:
		return common.ErrMemoryParityError
	case GatewayPathUnavailable:
		return common.ErrGatewayPathUnavailable
	case GatewayTargetDeviceFailedToRespond:
		return common.ErrGatewayTargetDeviceFailedToRespond
	default:
		return common.ErrUnknownException
	}
}

func ModbusOperationToBytes(operation ModbusOperation) []byte {
	if op, ok := operation.(ModbusWriteArrayRequest[[]bool]); ok {
		valueCount := len(op.Values())
		byteCount := getReturnByteCount(op.Values())
		data := make([]byte, 5+byteCount)
		data[0] = byte(op.Offset() >> 8)
		data[1] = byte(op.Offset())
		data[2] = byte(valueCount >> 8)
		data[3] = byte(valueCount)
		data[4] = byte(byteCount)
		for i, v := range op.Values() {
			if v {
				data[5+i/8] |= 1 << uint(i%8)
			}
		}
		return data
	} else if op, ok := operation.(ModbusWriteArrayRequest[[]uint16]); ok {
		valueCount := len(op.Values())
		byteCount := 2 * valueCount
		data := make([]byte, 5+byteCount)
		data[0] = byte(op.Offset() >> 8)
		data[1] = byte(op.Offset())
		data[2] = byte(valueCount >> 8)
		data[3] = byte(valueCount)
		data[4] = byte(byteCount)
		for i, v := range op.Values() {
			data[5+i*2] = byte(v >> 8)
			data[6+i*2] = byte(v)
		}
		return data
	} else if op, ok := operation.(ModbusWriteSingleRequest[bool]); ok {
		var valBytes []byte
		if op.Value() {
			valBytes = []byte{0xFF, 0x00}
		} else {
			valBytes = []byte{0x00, 0x00}
		}
		return []byte{
			byte(op.Offset() >> 8),
			byte(op.Offset()),
			valBytes[0],
			valBytes[1],
		}
	} else if op, ok := operation.(ModbusWriteSingleRequest[uint16]); ok {
		val := op.Value()
		valBytes := []byte{
			byte(val >> 8),
			byte(val),
		}
		return []byte{
			byte(op.Offset() >> 8),
			byte(op.Offset()),
			valBytes[0],
			valBytes[1],
		}
	} else if op, ok := operation.(ModbusWriteSingleResponse[bool]); ok {
		var valBytes []byte
		if op.Value() {
			valBytes = []byte{0xFF, 0x00}
		} else {
			valBytes = []byte{0x00, 0x00}
		}
		return []byte{
			byte(op.Offset() >> 8),
			byte(op.Offset()),
			valBytes[0],
			valBytes[1],
		}
	} else if op, ok := operation.(ModbusWriteSingleResponse[uint16]); ok {
		val := op.Value()
		valBytes := []byte{
			byte(val >> 8),
			byte(val),
		}
		return []byte{
			byte(op.Offset() >> 8),
			byte(op.Offset()),
			valBytes[0],
			valBytes[1],
		}
	} else if op, ok := operation.(ModbusWriteArrayResponse[[]uint16]); ok {
		return []byte{
			byte(op.Offset() >> 8),
			byte(op.Offset()),
			byte(op.Count() >> 8),
			byte(op.Count()),
		}
	} else if op, ok := operation.(ModbusWriteArrayResponse[[]bool]); ok {
		return []byte{
			byte(op.Offset() >> 8),
			byte(op.Offset()),
			byte(op.Count() >> 8),
			byte(op.Count()),
		}
	} else if op, ok := operation.(ModbusWriteSingleResponse[bool]); ok {
		var valBytes []byte
		if op.Value() {
			valBytes = []byte{0xFF, 0x00}
		} else {
			valBytes = []byte{0x00, 0x00}
		}
		return []byte{
			byte(op.Offset() >> 8),
			byte(op.Offset()),
			valBytes[0],
			valBytes[1],
		}
	} else if op, ok := operation.(ModbusWriteSingleResponse[uint16]); ok {
		val := op.Value()
		valBytes := []byte{
			byte(val >> 8),
			byte(val),
		}
		return []byte{
			byte(op.Offset() >> 8),
			byte(op.Offset()),
			valBytes[0],
			valBytes[1],
		}
	} else if op, ok := operation.(ModbusReadResponse[[]uint16]); ok {
		length := 2 * len(op.Values())
		data := make([]byte, 1+length)
		data[0] = byte(length)
		for i, v := range op.Values() {
			data[1+i*2] = byte(v >> 8)
			data[2+i*2] = byte(v)
		}
		return data
	} else if op, ok := operation.(ModbusReadResponse[[]bool]); ok {
		length := getReturnByteCount(op.Values())
		data := make([]byte, 1+length)
		data[0] = length
		for i, v := range op.Values() {
			if v {
				data[1+i/8] |= 1 << uint(i%8)
			}
		}
		return data
	} else if op, ok := operation.(ModbusReadRequest); ok {
		return []byte{
			byte(op.Offset() >> 8),
			byte(op.Offset()),
			byte(op.Count() >> 8),
			byte(op.Count()),
		}
	} else if op, ok := operation.(*ModbusOperationException); ok {
		return []byte{byte(op.ExceptionCode)}
	} else {
		panic("Unknown operation type")
	}
}
