package network

import (
	"sync"

	"github.com/rinzlerlabs/gomodbus/transport"
	"github.com/rinzlerlabs/gomodbus/transport/network"
)

func NewNetworkRequestCreator(newRequest func(transport.Header, *transport.ProtocolDataUnit) transport.ApplicationDataUnit) *networkRequestCreator {
	return &networkRequestCreator{
		transactionId: 0,
		newRequest:    newRequest,
	}
}

type networkRequestCreator struct {
	mu            sync.Mutex
	transactionId uint16
	newRequest    func(header transport.Header, pdu *transport.ProtocolDataUnit) transport.ApplicationDataUnit
}

func (s *networkRequestCreator) NewHeader(uint16) transport.Header {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.transactionId++
	txnId := []byte{byte(s.transactionId >> 8), byte(s.transactionId & 0xff)}
	return network.NewHeader(txnId, []byte{0x00, 0x00}, byte(0x01))
}

func (s *networkRequestCreator) NewRequest(header transport.Header, pdu *transport.ProtocolDataUnit) transport.ApplicationDataUnit {
	return s.newRequest(header, pdu)
}
