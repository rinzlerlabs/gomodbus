package network

func NewHeader(transactionid []byte, protocolid []byte, unitid byte) *header {
	return &header{transactionid: transactionid, protocolid: protocolid, unitid: unitid}
}

type header struct {
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
