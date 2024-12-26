package ascii

import (
	"context"
	"encoding/hex"

	"github.com/rinzlerlabs/gomodbus/common"
	"github.com/rinzlerlabs/gomodbus/data"
	"github.com/rinzlerlabs/gomodbus/transport"
	"github.com/rinzlerlabs/gomodbus/transport/serial"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type modbusApplicationDataUnit struct {
	header   transport.SerialHeader
	pdu      *transport.ProtocolDataUnit
	checksum transport.ErrorCheck
}

func (m modbusApplicationDataUnit) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddUint16("Address", uint16(m.header.Bytes()[0]))
	encoder.AddObject("PDU", m.pdu)
	return nil
}

func (m *modbusApplicationDataUnit) Header() transport.Header {
	return m.header
}

func (m *modbusApplicationDataUnit) PDU() *transport.ProtocolDataUnit {
	return m.pdu
}

func (m *modbusApplicationDataUnit) validateChecksum() error {
	for i, b := range m.Checksum() {
		if m.checksum[i] != b {
			return common.ErrInvalidChecksum
		}
	}
	return nil
}

func (m *modbusApplicationDataUnit) Checksum() transport.ErrorCheck {
	var lrc byte
	// address first
	lrc += byte(m.header.Bytes()[0])
	// then the data
	// TODO: Avoid the byte array allocation
	bytes := m.PDU().Bytes()
	for _, b := range bytes {
		lrc += b
	}
	// Two's complement
	lrc = ^lrc + 1
	return []byte{lrc}
}

func (m *modbusApplicationDataUnit) Bytes() []byte {
	return append(append(m.header.Bytes(), m.pdu.Bytes()...), m.Checksum()...)
}

func NewModbusRequestFrame(packet []byte) (*transport.ModbusFrame, error) {
	// parse the ascii to bytes
	packet = packet[1 : len(packet)-2] // remove the colon and the trailing CR LF
	packet, err := hex.DecodeString(string(packet))
	if err != nil {
		return nil, err
	}
	functionCode := data.FunctionCode(packet[1])
	op, err := data.ParseModbusRequestOperation(functionCode, packet[2:len(packet)-1])
	if err != nil {
		return nil, err
	}
	pdu := transport.NewProtocolDataUnit(op)
	adu := &modbusApplicationDataUnit{
		header:   serial.NewHeader(uint16(packet[0])),
		pdu:      pdu,
		checksum: packet[len(packet)-1:],
	}
	if adu.validateChecksum() != nil {
		return nil, common.ErrInvalidChecksum
	}
	return &transport.ModbusFrame{
		ApplicationDataUnit: adu,
		ResponseCreator:     NewModbusFrame,
	}, nil
}

func NewModbusFrame(header transport.Header, response *transport.ProtocolDataUnit) *transport.ModbusFrame {
	return &transport.ModbusFrame{
		ApplicationDataUnit: &modbusApplicationDataUnit{header: header.(transport.SerialHeader), pdu: response},
	}
}

func NewModbusASCIIResponseFrame(packet []byte, valueCount int) (*transport.ModbusFrame, error) {
	// parse the ascii to bytes
	packet = packet[1 : len(packet)-2] // remove the colon and the trailing CR LF
	packet, err := hex.DecodeString(string(packet))
	if err != nil {
		return nil, err
	}
	functionCode := data.FunctionCode(packet[1])
	op, err := data.ParseModbusResponseOperation(functionCode, packet[2:len(packet)-1], valueCount)
	if err != nil {
		return nil, err
	}
	pdu := transport.NewProtocolDataUnit(op)
	adu := &modbusApplicationDataUnit{
		header:   serial.NewHeader(uint16(packet[0])),
		pdu:      pdu,
		checksum: packet[len(packet)-1:],
	}
	if adu.validateChecksum() != nil {
		return nil, common.ErrInvalidChecksum
	}
	if pdu.FunctionCode().IsException() {
		return nil, pdu.Operation().(*data.ModbusOperationException).Error()
	}
	return &transport.ModbusFrame{
		ApplicationDataUnit: adu,
		ResponseCreator:     NewModbusFrame,
	}, nil
}

func NewModbusTransaction(frame *transport.ModbusFrame, t transport.Transport) transport.ModbusTransaction {
	return &modbusTransaction{
		frame:     frame,
		transport: t.(*modbusASCIITransport),
	}
}

type modbusTransaction struct {
	frame     *transport.ModbusFrame
	transport *modbusASCIITransport
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

	valueCount := 0
	if countable, success := m.frame.PDU().Operation().(data.CountableOperation); success {
		valueCount = countable.Count()
	}

	return NewModbusASCIIResponseFrame(b, valueCount)
}

func (m *modbusTransaction) Write(pdu *transport.ProtocolDataUnit) error {
	frame := m.frame.ResponseCreator(m.frame.Header(), pdu)
	m.transport.logger.Info("Response", zap.Object("Frame", frame))
	return m.transport.WriteFrame(frame)
}

func (m *modbusTransaction) Frame() *transport.ModbusFrame {
	return m.frame
}
