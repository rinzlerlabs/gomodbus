package network

import (
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
