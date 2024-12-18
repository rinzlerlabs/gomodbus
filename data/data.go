package data

import (
	"fmt"
	"strings"

	"github.com/rinzlerlabs/gomodbus/common"
	"go.uber.org/zap/zapcore"
)

type ModbusFrame interface {
	ApplicationDataUnit
	SendResponse(*ProtocolDataUnit) error
}

type ApplicationDataUnit interface {
	zapcore.ObjectMarshaler
	Bytes() []byte
	Address() uint16
	PDU() *ProtocolDataUnit
	Checksum() []byte
}

type ProtocolDataUnit struct {
	zapcore.ObjectMarshaler
	Function FunctionCode
	Data     []byte
}

func (pdu *ProtocolDataUnit) Bytes() []byte {
	// 1 byte for the function plus the data
	data := make([]byte, 1+len(pdu.Data))
	data[0] = byte(pdu.Function)
	copy(data[1:], pdu.Data)
	return data
}

func (pdu *ProtocolDataUnit) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddUint16("Function", uint16(pdu.Function))
	encoder.AddString("Data", EncodeToString(pdu.Data))
	return nil
}

func NewProtocolDataUnitFromBytes(data []byte) (*ProtocolDataUnit, error) {
	if len(data) < 2 {
		return nil, common.ErrInvalidPacket
	}
	return &ProtocolDataUnit{
		Function: FunctionCode(data[0]),
		Data:     data[1:],
	}, nil
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
