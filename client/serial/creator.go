package serial

import (
	"github.com/rinzlerlabs/gomodbus/transport"
	"github.com/rinzlerlabs/gomodbus/transport/serial"
)

func NewSerialRequestCreator(newHeader func(uint16) transport.Header, newRequest func(transport.Header, *transport.ProtocolDataUnit) (transport.ApplicationDataUnit, error)) *serialRequestCreator {
	return &serialRequestCreator{
		newHeader:  newHeader,
		newRequest: newRequest,
	}
}

type serialRequestCreator struct {
	newHeader  func(address uint16) transport.Header
	newRequest func(header transport.Header, pdu *transport.ProtocolDataUnit) (transport.ApplicationDataUnit, error)
}

func (s *serialRequestCreator) NewHeader(address uint16) transport.Header {
	return serial.NewHeader(address)
}

func (s *serialRequestCreator) NewRequest(header transport.Header, pdu *transport.ProtocolDataUnit) (transport.ApplicationDataUnit, error) {
	return s.newRequest(header, pdu)
}
