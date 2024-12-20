package serial

func NewHeader(address uint16) *header {
	return &header{address: address}
}

type header struct {
	address uint16
}

func (h *header) Address() uint16 {
	return h.address
}

func (h *header) Bytes() []byte {
	return []byte{byte(h.address)}
}
