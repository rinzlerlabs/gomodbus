package network

import (
	"sync"

	"github.com/rinzlerlabs/gomodbus/transport"
	"github.com/rinzlerlabs/gomodbus/transport/network"
)

func NewNetworkRequestCreator(createTransaction func(*transport.ModbusFrame, transport.Transport) transport.ModbusTransaction, newModbusFrame func(transport.Header, *transport.ProtocolDataUnit) *transport.ModbusFrame) *networkRequestCreator {
	return &networkRequestCreator{
		transactionId:     0,
		createTransaction: createTransaction,
		newModbusFrame:    newModbusFrame,
	}
}

type networkRequestCreator struct {
	mu                sync.Mutex
	transactionId     uint16
	createTransaction func(*transport.ModbusFrame, transport.Transport) transport.ModbusTransaction
	newModbusFrame    func(transport.Header, *transport.ProtocolDataUnit) *transport.ModbusFrame
}

func (s *networkRequestCreator) CreateTransaction(frame *transport.ModbusFrame, transport transport.Transport) transport.ModbusTransaction {
	return s.createTransaction(frame, transport)
}

func (s *networkRequestCreator) NewModbusFrame(address uint16, pdu *transport.ProtocolDataUnit) *transport.ModbusFrame {
	// we don't actually care about the address in TCP connections
	s.mu.Lock()
	defer s.mu.Unlock()
	s.transactionId++
	txnId := []byte{byte(s.transactionId >> 8), byte(s.transactionId & 0xff)}
	header := network.NewHeader(txnId, []byte{0x00, 0x00}, byte(0x01))
	return s.newModbusFrame(header, pdu)
}
