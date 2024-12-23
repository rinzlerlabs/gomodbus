package tcp

import (
	"context"
	"fmt"

	"github.com/rinzlerlabs/gomodbus/data"
	"github.com/rinzlerlabs/gomodbus/transport"
	"github.com/rinzlerlabs/gomodbus/transport/network"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type modbusApplicationDataUnit struct {
	header transport.TCPHeader
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
	return transport.ErrorCheck{0x00}
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

func NewModbusRequestFrame(packet []byte) (*transport.ModbusFrame, error) {
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
		header: network.NewHeader(txId, protoId, unitId),
		pdu:    pdu,
	}
	return &transport.ModbusFrame{
		ApplicationDataUnit: adu,
		ResponseCreator:     NewModbusFrame,
	}, nil
}

func NewModbusFrame(header transport.Header, response *transport.ProtocolDataUnit) *transport.ModbusFrame {
	return &transport.ModbusFrame{
		ApplicationDataUnit: &modbusApplicationDataUnit{
			header: header.(transport.TCPHeader),
			pdu:    response,
		},
	}
}

func NewModbusTCPResponseFrame(packet []byte, valueCount uint16) (*transport.ModbusFrame, error) {
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
		header: network.NewHeader(txId, protoId, unitId),
		pdu:    pdu,
	}
	return &transport.ModbusFrame{
		ApplicationDataUnit: adu,
		ResponseCreator:     NewModbusFrame,
	}, nil
}

func NewModbusTransaction(frame *transport.ModbusFrame, t transport.Transport) transport.ModbusTransaction {
	return &modbusTransaction{
		frame:     frame,
		transport: t.(*modbusTCPSocketTransport),
	}
}

type modbusTransaction struct {
	frame     *transport.ModbusFrame
	transport *modbusTCPSocketTransport
}

func (m *modbusTransaction) Read(ctx context.Context) (*transport.ModbusFrame, error) {
	return m.frame, nil
}

func (m *modbusTransaction) Exchange(ctx context.Context) (*transport.ModbusFrame, error) {
	err := m.transport.WriteFrame(m.frame)
	if err != nil {
		return nil, err
	}
	b, err := m.transport.readRawFrame(ctx)
	if err != nil {
		return nil, err
	}

	valueCount := uint16(0)
	if countable, success := m.frame.PDU().Operation().(data.CountableRequest); success {
		valueCount = countable.ValueCount()
	}

	return NewModbusTCPResponseFrame(b, valueCount)
}

func (m *modbusTransaction) Write(pdu *transport.ProtocolDataUnit) error {
	frame := m.frame.ResponseCreator(m.frame.Header(), pdu)
	m.transport.logger.Info("Response", zap.Object("Frame", frame))
	return m.transport.WriteFrame(frame)
}

func (m *modbusTransaction) Frame() *transport.ModbusFrame {
	return m.frame
}
