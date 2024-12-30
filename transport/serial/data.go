package serial

import (
	"sync"

	"github.com/rinzlerlabs/gomodbus/common"
	"github.com/rinzlerlabs/gomodbus/transport"
	"go.uber.org/zap/zapcore"
)

func NewHeader(address uint16) *header {
	return &header{address: address}
}

type header struct {
	zapcore.ObjectMarshaler
	address uint16
}

func (h *header) Address() uint16 {
	return h.address
}

func (h *header) Bytes() []byte {
	return []byte{byte(h.address)}
}

func (header header) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddUint16("Address", header.address)
	return nil
}

func NewFrameBuilder(aduCreator func(header transport.Header, pdu *transport.ProtocolDataUnit) (transport.ApplicationDataUnit, error)) *FrameBuilder {
	return &FrameBuilder{aduCreator: aduCreator}
}

type FrameBuilder struct {
	aduCreator func(header transport.Header, pdu *transport.ProtocolDataUnit) (transport.ApplicationDataUnit, error)
}

func (fb *FrameBuilder) BuildResponseFrame(header transport.Header, response *transport.ProtocolDataUnit) (transport.ApplicationDataUnit, error) {
	return fb.aduCreator(header, response)
}

type checksummer func(transport.ApplicationDataUnit) transport.ErrorCheck

func NewResponseModbusApplicationDataUnit(header transport.SerialHeader, pdu *transport.ProtocolDataUnit, checksummer checksummer) transport.ApplicationDataUnit {
	return &modbusApplicationDataUnit{
		header:      header,
		pdu:         pdu,
		checksummer: checksummer,
	}
}

func NewModbusApplicationDataUnit(header transport.SerialHeader, pdu *transport.ProtocolDataUnit, checksum transport.ErrorCheck, checksummer checksummer) (transport.ApplicationDataUnit, error) {
	adu := &modbusApplicationDataUnit{
		header:      header,
		pdu:         pdu,
		checksum:    checksum,
		checksummer: checksummer,
	}
	err := adu.validateChecksum()
	if err != nil {
		return nil, err
	}
	return adu, nil
}

type modbusApplicationDataUnit struct {
	mu          sync.Mutex
	header      transport.SerialHeader
	pdu         *transport.ProtocolDataUnit
	checksum    transport.ErrorCheck
	checksummer checksummer
}

func (m *modbusApplicationDataUnit) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
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
	for i, b := range m.checksummer(m) {
		if m.checksum[i] != b {
			return common.ErrInvalidChecksum
		}
	}
	return nil
}

func (m *modbusApplicationDataUnit) Checksum() transport.ErrorCheck {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.checksum == nil {
		m.checksum = m.checksummer(m)
	}
	return m.checksum
}

func (m *modbusApplicationDataUnit) Bytes() []byte {
	return append(append(m.header.Bytes(), m.pdu.Bytes()...), m.Checksum()...)
}
