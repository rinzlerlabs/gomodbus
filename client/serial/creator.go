package serial

import (
	"github.com/rinzlerlabs/gomodbus/transport"
	"github.com/rinzlerlabs/gomodbus/transport/serial"
)

func NewSerialRequestCreator(createTransaction func(*transport.ModbusFrame, transport.Transport) transport.ModbusTransaction, newModbusFrame func(transport.Header, *transport.ProtocolDataUnit) *transport.ModbusFrame) *serialRequestCreator {
	return &serialRequestCreator{
		createTransaction: createTransaction,
		newModbusFrame:    newModbusFrame,
	}
}

type serialRequestCreator struct {
	createTransaction func(*transport.ModbusFrame, transport.Transport) transport.ModbusTransaction
	newModbusFrame    func(transport.Header, *transport.ProtocolDataUnit) *transport.ModbusFrame
}

func (s *serialRequestCreator) CreateTransaction(frame *transport.ModbusFrame, transport transport.Transport) transport.ModbusTransaction {
	return s.createTransaction(frame, transport)
}

func (s *serialRequestCreator) NewModbusFrame(address uint16, pdu *transport.ProtocolDataUnit) *transport.ModbusFrame {
	header := serial.NewHeader(address)
	return s.newModbusFrame(header, pdu)
}
