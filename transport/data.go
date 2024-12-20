package transport

import (
	"context"
	"fmt"
	"strings"

	"github.com/rinzlerlabs/gomodbus/data"
	"go.uber.org/zap/zapcore"
)

type ErrorCheck []byte

type ModbusTransaction interface {
	Frame() *ModbusFrame
	Write(*ProtocolDataUnit) error
	Exchange(context.Context) (*ModbusFrame, error)
}

type ModbusFrame struct {
	zapcore.ObjectMarshaler
	ApplicationDataUnit
	ResponseCreator func(header Header, response *ProtocolDataUnit) *ModbusFrame
}

func (frame ModbusFrame) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddObject("Header", frame.Header())
	encoder.AddObject("PDU", frame.PDU())
	encoder.AddString("Checksum", EncodeToString(frame.Checksum()))
	return nil
}

type Header interface {
	zapcore.ObjectMarshaler
	Bytes() []byte
}

type SerialHeader interface {
	Header
	Address() uint16
}

type TCPHeader interface {
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

func NewProtocolDataUnit(functionCode data.FunctionCode, op data.ModbusOperation) *ProtocolDataUnit {
	return &ProtocolDataUnit{
		functionCode: functionCode,
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
	return append([]byte{byte(pdu.functionCode)}, pdu.op.Bytes()...)
}

type modbusFrame struct {
	ApplicationDataUnit
	transport       Transport
	responseCreator func(request ModbusFrame, response ProtocolDataUnit) ModbusFrame
}

func (m modbusFrame) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("Header", EncodeToString(m.Header().Bytes()))
	encoder.AddObject("PDU", m.PDU())
	encoder.AddString("Checksum", EncodeToString(m.Checksum()))
	return nil
}

func (m modbusFrame) Transport() Transport {
	return m.transport
}

func (m modbusFrame) ResponseCreator() func(request ModbusFrame, response ProtocolDataUnit) ModbusFrame {
	return m.responseCreator
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
