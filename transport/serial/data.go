package serial

import (
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

func NewFrameBuilder(aduCreator func(header transport.Header, pdu *transport.ProtocolDataUnit) transport.ApplicationDataUnit) *FrameBuilder {
	return &FrameBuilder{aduCreator: aduCreator}
}

type FrameBuilder struct {
	aduCreator func(header transport.Header, pdu *transport.ProtocolDataUnit) transport.ApplicationDataUnit
}

func (fb *FrameBuilder) BuildResponseFrame(header transport.Header, response *transport.ProtocolDataUnit) transport.ApplicationDataUnit {
	return fb.aduCreator(header, response)
}
