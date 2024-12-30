package network

import (
	"fmt"

	"github.com/rinzlerlabs/gomodbus/data"
	"github.com/rinzlerlabs/gomodbus/transport"
	"go.uber.org/zap/zapcore"
)

func NewHeader(transactionid []byte, protocolid []byte, unitid byte) *header {
	return &header{transactionid: transactionid, protocolid: protocolid, unitid: unitid}
}

type header struct {
	zapcore.ObjectMarshaler
	transactionid []byte
	protocolid    []byte
	unitid        byte
}

func (h header) TransactionID() []byte {
	return h.transactionid
}

func (h header) ProtocolID() []byte {
	return h.protocolid
}

func (h header) UnitID() byte {
	return h.unitid
}

func (h header) Bytes() []byte {
	bytes := make([]byte, 0)
	bytes = append(bytes, h.transactionid...)
	bytes = append(bytes, h.protocolid...)
	bytes = append(bytes, h.unitid)
	return bytes
}

func (header header) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("TransactionID", transport.EncodeToString(header.transactionid))
	encoder.AddString("ProtocolID", transport.EncodeToString(header.protocolid))
	encoder.AddString("UnitID", transport.EncodeToString([]byte{header.unitid}))
	return nil
}

type modbusApplicationDataUnit struct {
	header transport.NetworkHeader
	pdu    *transport.ProtocolDataUnit
}

func (m modbusApplicationDataUnit) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddObject("PDU", m.pdu)
	return nil
}

func (m *modbusApplicationDataUnit) Header() transport.Header {
	return m.header
}

func (m *modbusApplicationDataUnit) PDU() *transport.ProtocolDataUnit {
	return m.pdu
}

func (m *modbusApplicationDataUnit) Checksum() transport.ErrorCheck {
	return transport.ErrorCheck{}
}

func (m *modbusApplicationDataUnit) Bytes() []byte {
	bytes := make([]byte, 0)
	headerBytes := m.Header().Bytes()
	bytes = append(bytes, headerBytes[:4]...)
	pduBytes := m.pdu.Bytes()
	length := uint16(len(pduBytes))
	bytes = append(bytes, byte(length>>8), byte(length&0xFF))
	bytes = append(bytes, headerBytes[4:]...)
	bytes = append(bytes, pduBytes...)

	return bytes
}

func ParseModbusRequestFrame(packet []byte) (transport.ApplicationDataUnit, error) {
	txId := packet[0:2]
	protoId := packet[2:4]
	// length := packet[4:6]
	unitId := packet[6]
	packet = packet[7:]
	// TODO check length
	functionCode := data.FunctionCode(packet[0])
	op, err := data.ParseModbusRequestOperation(functionCode, packet[1:])
	if err != nil {
		return nil, err
	}
	pdu := transport.NewProtocolDataUnit(op)
	adu := &modbusApplicationDataUnit{
		header: NewHeader(txId, protoId, unitId),
		pdu:    pdu,
	}
	return adu, nil
}

func ParseModbusServerResponseFrame(packet []byte, valueCount int) (transport.ApplicationDataUnit, error) {
	txId := packet[0:2]
	protoId := packet[2:4]
	length := packet[4:6]
	unitId := packet[6]
	fmt.Printf("txId: %v, protoId: %v, length: %v, unitId: %v\n", txId, protoId, length, unitId)
	packet = packet[7:]
	functionCode := data.FunctionCode(packet[0])
	op, err := data.ParseModbusResponseOperation(functionCode, packet[1:], valueCount)
	if err != nil {
		return nil, err
	}
	pdu := transport.NewProtocolDataUnit(op)
	adu := &modbusApplicationDataUnit{
		header: NewHeader(txId, protoId, unitId),
		pdu:    pdu,
	}
	return adu, nil
}

func NewModbusResponse(header transport.Header, response *transport.ProtocolDataUnit) transport.ApplicationDataUnit {
	return &modbusApplicationDataUnit{header: header.(transport.NetworkHeader), pdu: response}
}

func NewModbusRequest(header transport.Header, response *transport.ProtocolDataUnit) transport.ApplicationDataUnit {
	return &modbusApplicationDataUnit{header: header.(transport.NetworkHeader), pdu: response}
}

func NewFrameBuilder() transport.FrameBuilder {
	return &frameBuilder{}
}

type frameBuilder struct{}

func (fb *frameBuilder) BuildResponseFrame(header transport.Header, response *transport.ProtocolDataUnit) transport.ApplicationDataUnit {
	return &modbusApplicationDataUnit{header: header.(transport.NetworkHeader), pdu: response}
}
