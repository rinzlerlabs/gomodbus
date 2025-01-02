package rtu

import (
	"github.com/rinzlerlabs/gomodbus/common"
	"github.com/rinzlerlabs/gomodbus/data"
	"github.com/rinzlerlabs/gomodbus/transport"
	"github.com/rinzlerlabs/gomodbus/transport/serial"
)

func NewModbusApplicationDataUnit(header transport.Header, pdu *transport.ProtocolDataUnit) (transport.ApplicationDataUnit, error) {
	if serialHeader, ok := header.(transport.SerialHeader); ok {
		return serial.NewResponseModbusApplicationDataUnit(serialHeader, pdu, checksummer), nil
	}
	return nil, common.ErrInvalidHeader
}

func checksummer(m transport.ApplicationDataUnit) transport.ErrorCheck {
	var crc uint16 = 0xFFFF
	// TODO: avoid the byte array allocation
	bytes := make([]byte, 0)
	bytes = append(bytes, m.Header().Bytes()...)
	bytes = append(bytes, byte(m.PDU().FunctionCode()))
	bytes = append(bytes, data.ModbusOperationToBytes(m.PDU().Operation())...)
	for _, b := range bytes {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if (crc & 1) != 0 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return []byte{byte(crc), byte(crc >> 8)}
}

func newModbusRTURequestProtocolDataUnit(functionCode data.FunctionCode, bytes []byte) (*transport.ProtocolDataUnit, error) {
	op, err := data.ParseModbusRequestOperation(functionCode, bytes)
	return transport.NewProtocolDataUnit(op), err
}

// ParseModbusRequestFrame creates a new Modbus RTU request frame from raw bytes read from the wire and a transport
func ParseModbusRequestFrame(packet []byte) (transport.ApplicationDataUnit, error) {
	pdu, err := newModbusRTURequestProtocolDataUnit(data.FunctionCode(packet[1]), packet[2:len(packet)-2])
	if err != nil {
		return nil, err
	}
	return serial.NewModbusApplicationDataUnit(serial.NewHeader(uint16(packet[0])), pdu, packet[len(packet)-2:], checksummer)
}

func ParseModbusResponseFrame(packet []byte, valueCount int) (transport.ApplicationDataUnit, error) {
	functionCode := data.FunctionCode(packet[1])
	op, err := data.ParseModbusResponseOperation(functionCode, packet[2:len(packet)-2], valueCount)
	if err != nil {
		return nil, err
	}
	pdu := transport.NewProtocolDataUnit(op)
	adu, err := serial.NewModbusApplicationDataUnit(serial.NewHeader(uint16(packet[0])), pdu, packet[len(packet)-2:], checksummer)
	if err != nil {
		return nil, err
	}
	if pdu.FunctionCode().IsException() {
		return nil, pdu.Operation().(*data.ModbusOperationException).Error()
	}
	return adu, nil
}
