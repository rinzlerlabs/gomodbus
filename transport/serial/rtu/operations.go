package rtu

import (
	"github.com/rinzlerlabs/gomodbus/common"
	"github.com/rinzlerlabs/gomodbus/data"
	"github.com/rinzlerlabs/gomodbus/transport"
	"go.uber.org/zap/zapcore"
)

type RTUApplicationDataUnit struct {
	zapcore.ObjectMarshaler
	address uint16
	pdu     *data.ProtocolDataUnit
}

func (adu *RTUApplicationDataUnit) Bytes() []byte {
	pdu := adu.pdu.Bytes()
	// 1 byte for the address
	data := make([]byte, len(pdu)+1)
	data[0] = byte(adu.address)
	copy(data[1:], pdu)
	data = append(data, adu.Checksum()...)
	return data
}

func (adu *RTUApplicationDataUnit) Address() uint16 {
	return adu.address
}

func (adu *RTUApplicationDataUnit) PDU() *data.ProtocolDataUnit {
	return adu.pdu
}

func (adu *RTUApplicationDataUnit) Checksum() []byte {
	var crc uint16 = 0xFFFF
	// TODO: avoid the byte array allocation
	data := make([]byte, 1+1+len(adu.pdu.Data))
	data[0] = byte(adu.address)
	data[1] = byte(adu.pdu.Function)
	copy(data[2:], adu.pdu.Data)
	for _, b := range data {
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

func (adu *RTUApplicationDataUnit) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddUint16("Address", adu.address)
	encoder.AddObject("PDU", adu.pdu)
	return nil
}

func NewApplicationDataUnitFromRequest(address uint16, pdu *data.ProtocolDataUnit) data.ModbusFrame {
	return &ModbusRTUFrame{
		RTUApplicationDataUnit: &RTUApplicationDataUnit{
			address: address,
			pdu:     pdu,
		},
		transport: nil,
	}
}

func NewApplicationDataUnitFromWire(d []byte) (*RTUApplicationDataUnit, error) {
	if len(d) < 4 {
		return nil, common.ErrInvalidPacket
	}

	checksum := uint16(d[len(d)-2])<<8 | uint16(d[len(d)-1])

	// 1 because 1 byte for the address
	// -2 because of the checksum
	pdu, err := data.NewProtocolDataUnitFromBytes(d[1 : len(d)-2])
	if err != nil {
		return nil, err
	}
	adu := &RTUApplicationDataUnit{
		address: uint16(d[0]),
		pdu:     pdu,
	}

	err = validateCRCChecksum(checksum, adu)
	if err != nil {
		return nil, err
	}

	return adu, nil
}

func validateCRCChecksum(wireChecksum uint16, adu *RTUApplicationDataUnit) error {
	// the last two bytes are the CRC
	data := adu.Checksum()
	crc := uint16(data[len(data)-2])<<8 | uint16(data[len(data)-1])
	if crc != wireChecksum {
		return common.ErrInvalidChecksum
	}
	return nil
}

func NewModbusFrame(data []byte, transport transport.Transport) (data.ModbusFrame, error) {
	adu, err := NewApplicationDataUnitFromWire(data)
	if err != nil {
		return nil, err
	}
	return &ModbusRTUFrame{
		RTUApplicationDataUnit: adu,
		transport:              transport,
	}, nil
}

type ModbusRTUFrame struct {
	*RTUApplicationDataUnit
	transport transport.Transport
}

func (f *ModbusRTUFrame) SendResponse(pdu *data.ProtocolDataUnit) error {
	adu := NewApplicationDataUnitFromRequest(f.address, pdu)
	return f.transport.WriteFrame(adu)
}
