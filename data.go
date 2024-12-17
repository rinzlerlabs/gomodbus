package gomodbus

import (
	"encoding/hex"
	"strconv"
)

type ApplicationDataUnit interface {
	Bytes() []byte
	Address() uint16
	PDU() *ProtocolDataUnit
	Checksum() []byte
}

type RTUApplicationDataUnit struct {
	address uint16
	pdu     *ProtocolDataUnit
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

func (adu *RTUApplicationDataUnit) PDU() *ProtocolDataUnit {
	return adu.pdu
}

func (adu *RTUApplicationDataUnit) Checksum() []byte {
	var crc uint16 = 0xFFFF
	// TODO: avoid the byte array allocation
	data := make([]byte, 1+1+len(adu.pdu.Data))
	data[0] = byte(adu.address)
	data[1] = adu.pdu.Function
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

type ASCIIApplicationDataUnit struct {
	address uint16
	pdu     *ProtocolDataUnit
}

func (adu *ASCIIApplicationDataUnit) Bytes() []byte {
	pdu := adu.pdu.Bytes()
	// 1 byte for the address
	data := make([]byte, len(pdu)+1)
	data[0] = byte(adu.address)
	copy(data[1:], pdu)
	data = append(data, adu.Checksum()...)
	return data
}

func (adu *ASCIIApplicationDataUnit) Address() uint16 {
	return adu.address
}

func (adu *ASCIIApplicationDataUnit) PDU() *ProtocolDataUnit {
	return adu.pdu
}

func (adu *ASCIIApplicationDataUnit) Checksum() []byte {
	var lrc byte
	// address first
	lrc += byte(adu.address)
	// then the function code
	lrc += adu.pdu.Function
	// then the data
	for _, b := range adu.pdu.Data {
		lrc += b
	}
	// Two's complement
	lrc = ^lrc + 1
	return []byte{lrc}
}

func NewASCIIApplicationDataUnitFromResponse(address uint16, function byte, response ModbusResponse) ApplicationDataUnit {
	pdu := &ProtocolDataUnit{
		Function: function,
		Data:     response.Bytes(),
	}
	return &ASCIIApplicationDataUnit{
		address: address,
		pdu:     pdu,
	}
}

func NewASCIIApplicationDataUnitFromRequest(data string) (ApplicationDataUnit, error) {
	if len(data) < 6 {
		return nil, ErrInvalidPacket
	}
	if data[0] != ':' || data[len(data)-2] != '\r' || data[len(data)-1] != '\n' {
		return nil, ErrInvalidPacket
	}
	// ditch the newline characters
	data = data[1 : len(data)-2]
	dataWithoutChecksum, err := hex.DecodeString(data[:len(data)-2])
	if err != nil {
		return nil, err
	}

	checksum, err := strconv.ParseUint(data[len(data)-2:], 16, 16)
	if err != nil {
		return nil, err
	}

	pdu, err := NewProtocolDataUnitFromBytes(dataWithoutChecksum[1:])
	if err != nil {
		return nil, err
	}

	adu := &ASCIIApplicationDataUnit{
		address: uint16(dataWithoutChecksum[0]),
		pdu:     pdu,
	}

	err = validateLRCChecksum(uint16(checksum), adu)
	if err != nil {
		return nil, err
	}

	return adu, nil
}

func NewRTUApplicationDataUnitFromResponse(address uint16, function byte, response ModbusResponse) ApplicationDataUnit {
	pdu := &ProtocolDataUnit{
		Function: function,
		Data:     response.Bytes(),
	}
	return &RTUApplicationDataUnit{
		address: address,
		pdu:     pdu,
	}
}

func NewRTUApplicationDataUnitFromRequest(data []byte) (ApplicationDataUnit, error) {
	if len(data) < 4 {
		return nil, ErrInvalidPacket
	}

	checksum := uint16(data[len(data)-2])<<8 | uint16(data[len(data)-1])

	// 1 because 1 byte for the address
	// -2 because of the checksum
	pdu, err := NewProtocolDataUnitFromBytes(data[1 : len(data)-2])
	if err != nil {
		return nil, err
	}
	adu := &RTUApplicationDataUnit{
		address: uint16(data[0]),
		pdu:     pdu,
	}

	err = validateCRCChecksum(checksum, adu)
	if err != nil {
		return nil, err
	}

	return adu, nil
}

func validateCRCChecksum(wireChecksum uint16, adu ApplicationDataUnit) error {
	// the last two bytes are the CRC
	data := adu.Checksum()
	crc := uint16(data[len(data)-2])<<8 | uint16(data[len(data)-1])
	if crc != wireChecksum {
		return ErrInvalidChecksum
	}
	return nil
}

type ProtocolDataUnit struct {
	Function byte
	Data     []byte
}

func (pdu *ProtocolDataUnit) Bytes() []byte {
	// 1 byte for the function plus the data
	data := make([]byte, 1+len(pdu.Data))
	data[0] = pdu.Function
	copy(data[1:], pdu.Data)
	return data
}

func NewProtocolDataUnitFromBytes(data []byte) (*ProtocolDataUnit, error) {
	if len(data) < 2 {
		return nil, ErrInvalidPacket
	}
	return &ProtocolDataUnit{
		Function: data[0],
		Data:     data[1:],
	}, nil
}
