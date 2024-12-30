package transport

import (
	"fmt"
	"strings"

	"github.com/rinzlerlabs/gomodbus/data"
	"go.uber.org/zap/zapcore"
)

type ErrorCheck []byte

type Header interface {
	zapcore.ObjectMarshaler
	Bytes() []byte
}

type SerialHeader interface {
	Header
	Address() uint16
}

type NetworkHeader interface {
	Header
	TransactionID() []byte
	ProtocolID() []byte
	UnitID() byte
}

type ApplicationDataUnit interface {
	zapcore.ObjectMarshaler
	Bytes() []byte
	Header() Header
	PDU() *ProtocolDataUnit
	Checksum() ErrorCheck
}

func NewProtocolDataUnit(op data.ModbusOperation) *ProtocolDataUnit {
	var f data.FunctionCode
	switch op := op.(type) {
	case *data.ReadCoilsRequest:
		f = data.ReadCoils
	case *data.ReadCoilsResponse:
		f = data.ReadCoils
	case *data.ReadDiscreteInputsRequest:
		f = data.ReadDiscreteInputs
	case *data.ReadDiscreteInputsResponse:
		f = data.ReadDiscreteInputs
	case *data.ReadHoldingRegistersRequest:
		f = data.ReadHoldingRegisters
	case *data.ReadHoldingRegistersResponse:
		f = data.ReadHoldingRegisters
	case *data.ReadInputRegistersRequest:
		f = data.ReadInputRegisters
	case *data.ReadInputRegistersResponse:
		f = data.ReadInputRegisters
	case *data.WriteSingleCoilRequest:
		f = data.WriteSingleCoil
	case *data.WriteSingleCoilResponse:
		f = data.WriteSingleCoil
	case *data.WriteSingleRegisterRequest:
		f = data.WriteSingleRegister
	case *data.WriteSingleRegisterResponse:
		f = data.WriteSingleRegister
	case *data.WriteMultipleCoilsRequest:
		f = data.WriteMultipleCoils
	case *data.WriteMultipleCoilsResponse:
		f = data.WriteMultipleCoils
	case *data.WriteMultipleRegistersRequest:
		f = data.WriteMultipleRegisters
	case *data.WriteMultipleRegistersResponse:
		f = data.WriteMultipleRegisters
	case *data.ModbusOperationException:
		f = op.FunctionCode
	}
	return &ProtocolDataUnit{
		functionCode: f,
		op:           op,
	}
}

type ProtocolDataUnit struct {
	functionCode data.FunctionCode
	op           data.ModbusOperation
}

func (pdu *ProtocolDataUnit) Operation() data.ModbusOperation {
	return pdu.op
}

func (pdu ProtocolDataUnit) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddUint16("Function", uint16(pdu.functionCode))
	encoder.AddObject("Operation", pdu.op)
	return nil
}

func (pdu *ProtocolDataUnit) FunctionCode() data.FunctionCode {
	return pdu.functionCode
}

func (pdu *ProtocolDataUnit) Bytes() []byte {
	return append([]byte{byte(pdu.functionCode)}, data.ModbusOperationToBytes(pdu.op)...)
}

func EncodeToString(data []byte) string {
	var builder strings.Builder
	for i, b := range data {
		if i > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString(fmt.Sprintf("%02X", b))
	}
	return builder.String()
}
