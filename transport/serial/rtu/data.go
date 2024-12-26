package rtu

import (
	"context"

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
	encoder.AddUint16("Address", m.header.Address())
	encoder.AddObject("PDU", m.pdu)
	return nil
}

func (a *modbusApplicationDataUnit) Header() transport.Header {
	return a.header
}

func (a *modbusApplicationDataUnit) PDU() *transport.ProtocolDataUnit {
	return a.pdu
}

func (a *modbusApplicationDataUnit) validateChecksum() error {
	for i, b := range a.Checksum() {
		if a.checksum[i] != b {
			return common.ErrInvalidChecksum
		}
	}
	return nil
}

func (a *modbusApplicationDataUnit) Checksum() transport.ErrorCheck {
	var crc uint16 = 0xFFFF
	// TODO: avoid the byte array allocation
	bytes := make([]byte, 0)
	bytes = append(bytes, a.header.Bytes()...)
	bytes = append(bytes, byte(a.pdu.FunctionCode()))
	bytes = append(bytes, data.ModbusOperationToBytes(a.pdu.Operation())...)
	for _, b := range bytes {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if (crc & 1) != 0 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return []byte{byte(crc), byte(crc >> 8)}
}

func (a *modbusApplicationDataUnit) Bytes() []byte {
	return append(append(a.header.Bytes(), a.pdu.Bytes()...), a.Checksum()...)
}

func newModbusRTURequestProtocolDataUnit(functionCode data.FunctionCode, bytes []byte) (*transport.ProtocolDataUnit, error) {
	op, err := data.ParseModbusRequestOperation(functionCode, bytes)
	return transport.NewProtocolDataUnit(op), err
}

// NewModbusRTURequestFrame creates a new Modbus RTU request frame from raw bytes read from the wire and a transport
func NewModbusRTURequestFrame(packet []byte) (*transport.ModbusFrame, error) {
	pdu, err := newModbusRTURequestProtocolDataUnit(data.FunctionCode(packet[1]), packet[2:len(packet)-2])
	if err != nil {
		return nil, err
	}
	adu := &modbusApplicationDataUnit{
		header:   serial.NewHeader(uint16(packet[0])),
		pdu:      pdu,
		checksum: packet[len(packet)-2:],
	}
	if adu.validateChecksum() != nil {
		return nil, common.ErrInvalidChecksum
	}
	return &transport.ModbusFrame{
		ApplicationDataUnit: adu,
		ResponseCreator:     NewModbusFrame,
	}, nil
}

// NewModbusRTUResponseFramer creates a new Modbus RTU response frame from a request frame and a response PDU
func NewModbusFrame(header transport.Header, response *transport.ProtocolDataUnit) *transport.ModbusFrame {
	return &transport.ModbusFrame{
		ApplicationDataUnit: &modbusApplicationDataUnit{header: header.(transport.SerialHeader), pdu: response},
	}
}

func NewModbusRTUResponseFrame(packet []byte, valueCount int) (*transport.ModbusFrame, error) {
	functionCode := data.FunctionCode(packet[1])
	op, err := data.ParseModbusResponseOperation(functionCode, packet[2:len(packet)-2], valueCount)
	if err != nil {
		return nil, err
	}
	pdu := transport.NewProtocolDataUnit(op)
	adu := &modbusApplicationDataUnit{
		header:   serial.NewHeader(uint16(packet[0])),
		pdu:      pdu,
		checksum: packet[len(packet)-2:],
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
		transport: t.(*modbusRTUTransport),
	}
}

type modbusTransaction struct {
	frame     *transport.ModbusFrame
	transport *modbusRTUTransport
}

func (m *modbusTransaction) Read(ctx context.Context) (*transport.ModbusFrame, error) {
	return m.frame, nil
}

func (m *modbusTransaction) Exchange(ctx context.Context) (*transport.ModbusFrame, error) {
	err := m.transport.WriteFrame(m.frame)
	if err != nil {
		return nil, err
	}
	b, err := m.transport.readResponseFrame(ctx)
	if err != nil {
		return nil, err
	}

	valueCount := 0
	if countable, success := m.frame.PDU().Operation().(data.CountableOperation); success {
		valueCount = countable.Count()
	}
	return NewModbusRTUResponseFrame(b, valueCount)
}

func (m *modbusTransaction) Write(pdu *transport.ProtocolDataUnit) error {
	frame := m.frame.ResponseCreator(m.frame.Header(), pdu)
	m.transport.logger.Info("Response", zap.Object("Frame", frame))
	return m.transport.WriteFrame(frame)
}

func (m *modbusTransaction) Frame() *transport.ModbusFrame {
	return m.frame
}
