package serial

import "go.uber.org/zap/zapcore"

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
