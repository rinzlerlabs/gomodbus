package ascii

import (
	"encoding/hex"
	"errors"

	"github.com/rinzlerlabs/gomodbus/data"
	"github.com/rinzlerlabs/gomodbus/transport"
	"github.com/rinzlerlabs/gomodbus/transport/serial"
)

func NewModbusApplicationDataUnit(header transport.Header, pdu *transport.ProtocolDataUnit) (transport.ApplicationDataUnit, error) {
	if serialHeader, ok := header.(transport.SerialHeader); ok {
		return serial.NewResponseModbusApplicationDataUnit(serialHeader, pdu, checksummer), nil
	}
	return nil, errors.New("invalid header")
}

func checksummer(m transport.ApplicationDataUnit) transport.ErrorCheck {
	var lrc byte
	// address first
	lrc += byte(m.Header().Bytes()[0])
	// then the data
	// TODO: Avoid the byte array allocation
	bytes := m.PDU().Bytes()
	for _, b := range bytes {
		lrc += b
	}
	// Two's complement
	lrc = ^lrc + 1
	return []byte{lrc}
}

func ParseModbusRequestFrame(packet []byte) (transport.ApplicationDataUnit, error) {
	// parse the ascii to bytes
	packet = packet[1 : len(packet)-2] // remove the colon and the trailing CR LF
	packet, err := hex.DecodeString(string(packet))
	if err != nil {
		return nil, err
	}
	functionCode := data.FunctionCode(packet[1])
	op, err := data.ParseModbusRequestOperation(functionCode, packet[2:len(packet)-1])
	if err != nil {
		return nil, err
	}
	pdu := transport.NewProtocolDataUnit(op)
	return serial.NewModbusApplicationDataUnit(serial.NewHeader(uint16(packet[0])), pdu, packet[len(packet)-1:], checksummer)
}

func ParseModbusResponseFrame(packet []byte, valueCount int) (transport.ApplicationDataUnit, error) {
	// parse the ascii to bytes
	packet = packet[1 : len(packet)-2] // remove the colon and the trailing CR LF
	packet, err := hex.DecodeString(string(packet))
	if err != nil {
		return nil, err
	}
	functionCode := data.FunctionCode(packet[1])
	op, err := data.ParseModbusResponseOperation(functionCode, packet[2:len(packet)-1], valueCount)
	if err != nil {
		return nil, err
	}
	pdu := transport.NewProtocolDataUnit(op)
	adu, err := serial.NewModbusApplicationDataUnit(serial.NewHeader(uint16(packet[0])), pdu, packet[len(packet)-1:], checksummer)
	if err != nil {
		return nil, err
	}
	// check if the pdu is an exception returned by the server
	if pdu.FunctionCode().IsException() {
		return nil, pdu.Operation().(*data.ModbusOperationException).Error()
	}
	return adu, nil
}
