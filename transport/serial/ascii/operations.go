package ascii

import (
	"encoding/hex"
	"strconv"

	"github.com/rinzlerlabs/gomodbus/common"
	"github.com/rinzlerlabs/gomodbus/data"
	"github.com/rinzlerlabs/gomodbus/transport"
	"go.uber.org/zap/zapcore"
)

type ASCIIApplicationDataUnit struct {
	zapcore.ObjectMarshaler
	address uint16
	pdu     *data.ProtocolDataUnit
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

func (adu *ASCIIApplicationDataUnit) PDU() *data.ProtocolDataUnit {
	return adu.pdu
}

func (adu *ASCIIApplicationDataUnit) Checksum() []byte {
	var lrc byte
	// address first
	lrc += byte(adu.address)
	// then the function code
	lrc += byte(adu.pdu.Function)
	// then the data
	for _, b := range adu.pdu.Data {
		lrc += b
	}
	// Two's complement
	lrc = ^lrc + 1
	return []byte{lrc}
}

func (adu *ASCIIApplicationDataUnit) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddUint16("Address", adu.address)
	encoder.AddObject("PDU", adu.pdu)
	return nil
}

func NewApplicationDataUnitFromRequest(address uint16, pdu *data.ProtocolDataUnit) data.ModbusFrame {
	return &ModbusASCIIFrame{
		ASCIIApplicationDataUnit: &ASCIIApplicationDataUnit{
			address: address,
			pdu:     pdu,
		},
		transport: nil,
	}
}

func NewApplicationDataUnitFromWire(d string) (*ASCIIApplicationDataUnit, error) {
	if len(d) < 6 {
		return nil, common.ErrInvalidPacket
	}
	if d[0] != ':' || d[len(d)-2] != '\r' || d[len(d)-1] != '\n' {
		return nil, common.ErrInvalidPacket
	}
	// ditch the newline characters
	d = d[1 : len(d)-2]
	dataWithoutChecksum, err := hex.DecodeString(d[:len(d)-2])
	if err != nil {
		return nil, err
	}

	checksum, err := strconv.ParseUint(d[len(d)-2:], 16, 16)
	if err != nil {
		return nil, err
	}

	pdu, err := data.NewProtocolDataUnitFromBytes(dataWithoutChecksum[1:])
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

func validateLRCChecksum(wireChecksum uint16, adu *ASCIIApplicationDataUnit) error {
	lrc := adu.Checksum()
	if len(lrc) != 1 {
		return common.ErrInvalidChecksum
	}
	if lrc[0] != byte(wireChecksum) {
		return common.ErrInvalidChecksum
	}
	return nil
}

func NewModbusFrame(data []byte, transport transport.Transport) (data.ModbusFrame, error) {
	adu, err := NewApplicationDataUnitFromWire(string(data))
	if err != nil {
		return nil, err
	}
	return &ModbusASCIIFrame{
		ASCIIApplicationDataUnit: adu,
		transport:                transport,
	}, nil
}

type ModbusASCIIFrame struct {
	*ASCIIApplicationDataUnit
	transport transport.Transport
}

func (f *ModbusASCIIFrame) SendResponse(pdu *data.ProtocolDataUnit) error {
	adu := NewApplicationDataUnitFromRequest(f.address, pdu)
	return f.transport.WriteFrame(adu)
}
