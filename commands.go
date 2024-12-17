package gomodbus

const (
	FunctionCodeReadCoils              byte = 0x01
	FunctionCodeReadDiscreteInputs     byte = 0x02
	FunctionCodeReadHoldingRegisters   byte = 0x03
	FunctionCodeReadInputRegisters     byte = 0x04
	FunctionCodeWriteSingleCoil        byte = 0x05
	FunctionCodeWriteSingleRegister    byte = 0x06
	FunctionCodeWriteMultipleCoils     byte = 0x0F
	FunctionCodeWriteMultipleRegisters byte = 0x10
)

type ModbusResponse interface {
	Bytes() []byte
}

type ReadCoilsRequest struct {
	Offset uint16
	Count  uint16
}

func NewReadCoilsRequest(adu ApplicationDataUnit) (*ReadCoilsRequest, error) {
	pdu := adu.PDU()
	if len(pdu.Data) != 4 {
		return nil, ErrInvalidPacket
	}
	return &ReadCoilsRequest{
		Offset: uint16(pdu.Data[0])<<8 | uint16(pdu.Data[1]),
		Count:  uint16(pdu.Data[2])<<8 | uint16(pdu.Data[3]),
	}, nil
}

type ReadCoilsResponse struct {
	Values []bool
}

func (r *ReadCoilsResponse) Bytes() []byte {
	length := getReturnByteCount(r.Values)
	data := make([]byte, 1+length)
	data[0] = length
	for i, v := range r.Values {
		if v {
			data[1+i/8] |= 1 << uint(i%8)
		}
	}
	return data
}

func getReturnByteCount(values []bool) byte {
	if len(values)%8 == 0 {
		return byte(len(values) / 8)
	}
	return byte(len(values)/8 + 1)
}

type ReadDiscreteInputsRequest struct {
	Offset uint16
	Count  uint16
}

func NewReadDiscreteInputsRequest(adu ApplicationDataUnit) (*ReadDiscreteInputsRequest, error) {
	pdu := adu.PDU()
	if len(pdu.Data) != 4 {
		return nil, ErrInvalidPacket
	}
	return &ReadDiscreteInputsRequest{
		Offset: uint16(pdu.Data[0])<<8 | uint16(pdu.Data[1]),
		Count:  uint16(pdu.Data[2])<<8 | uint16(pdu.Data[3]),
	}, nil
}

type ReadDiscreteInputsResponse struct {
	Values []bool
}

func (r *ReadDiscreteInputsResponse) Bytes() []byte {
	length := getReturnByteCount(r.Values)
	data := make([]byte, 1+length)
	data[0] = length
	for i, v := range r.Values {
		if v {
			data[1+i/8] |= 1 << uint(i%8)
		}
	}
	return data
}

type ReadHoldingRegistersRequest struct {
	Offset uint16
	Count  uint16
}

func NewReadHoldingRegistersRequest(adu ApplicationDataUnit) (*ReadHoldingRegistersRequest, error) {
	pdu := adu.PDU()
	if len(pdu.Data) != 4 {
		return nil, ErrInvalidPacket
	}
	return &ReadHoldingRegistersRequest{
		Offset: uint16(pdu.Data[0])<<8 | uint16(pdu.Data[1]),
		Count:  uint16(pdu.Data[2])<<8 | uint16(pdu.Data[3]),
	}, nil
}

type ReadHoldingRegistersResponse struct {
	Values []uint16
}

func (r *ReadHoldingRegistersResponse) Bytes() []byte {
	length := 2 * len(r.Values)
	data := make([]byte, 1+length)
	data[0] = byte(length)
	for i, v := range r.Values {
		data[1+i*2] = byte(v >> 8)
		data[2+i*2] = byte(v)
	}
	return data
}

type ReadInputRegistersRequest struct {
	Offset uint16
	Count  uint16
}

func NewReadInputRegistersRequest(adu ApplicationDataUnit) (*ReadInputRegistersRequest, error) {
	pdu := adu.PDU()
	if len(pdu.Data) != 4 {
		return nil, ErrInvalidPacket
	}
	return &ReadInputRegistersRequest{
		Offset: uint16(pdu.Data[0])<<8 | uint16(pdu.Data[1]),
		Count:  uint16(pdu.Data[2])<<8 | uint16(pdu.Data[3]),
	}, nil
}

type ReadInputRegistersResponse struct {
	Values []uint16
}

func (r *ReadInputRegistersResponse) Bytes() []byte {
	length := 2 * len(r.Values)
	data := make([]byte, 1+length)
	data[0] = byte(length)
	for i, v := range r.Values {
		data[1+i*2] = byte(v >> 8)
		data[2+i*2] = byte(v)
	}
	return data
}

type WriteSingleCoilRequest struct {
	Offset uint16
	Value  bool
}

func NewWriteSingleCoilRequest(adu ApplicationDataUnit) (*WriteSingleCoilRequest, error) {
	pdu := adu.PDU()
	if len(pdu.Data) != 4 {
		return nil, ErrInvalidPacket
	}
	return &WriteSingleCoilRequest{
		Offset: uint16(pdu.Data[0])<<8 | uint16(pdu.Data[1]),
		Value:  pdu.Data[2] == 0x00 && pdu.Data[3] == 0xFF,
	}, nil
}

type WriteSingleCoilResponse struct {
	Offset uint16
	Value  bool
}

func (r *WriteSingleCoilResponse) Bytes() []byte {
	val := byte(0x00)
	if r.Value {
		val = 0xFF
	}
	return []byte{
		byte(r.Offset >> 8),
		byte(r.Offset),
		0x00,
		val,
	}
}

type WriteSingleRegisterRequest struct {
	Offset uint16
	Value  uint16
}

func NewWriteSingleRegisterRequest(adu ApplicationDataUnit) (*WriteSingleRegisterRequest, error) {
	pdu := adu.PDU()
	if len(pdu.Data) != 4 {
		return nil, ErrInvalidPacket
	}
	return &WriteSingleRegisterRequest{
		Offset: uint16(pdu.Data[0])<<8 | uint16(pdu.Data[1]),
		Value:  uint16(pdu.Data[2])<<8 | uint16(pdu.Data[3]),
	}, nil
}

type WriteSingleRegisterResponse struct {
	Offset uint16
	Value  uint16
}

func (r *WriteSingleRegisterResponse) Bytes() []byte {
	return []byte{
		byte(r.Offset >> 8),
		byte(r.Offset),
		byte(r.Value >> 8),
		byte(r.Value),
	}
}

type WriteMultipleCoilsRequest struct {
	Offset uint16
	Values []bool
}

func NewWriteMultipleCoilsRequest(adu ApplicationDataUnit) (*WriteMultipleCoilsRequest, error) {
	pdu := adu.PDU()
	if len(pdu.Data) < 5 {
		return nil, ErrInvalidPacket
	}
	offset := uint16(pdu.Data[0])<<8 | uint16(pdu.Data[1])
	coilCount := uint16(pdu.Data[2])<<8 | uint16(pdu.Data[3])
	byteCount := uint16(pdu.Data[4])
	if uint16(len(pdu.Data)) != 2+2+byteCount+1 {
		return nil, ErrInvalidPacket
	}
	if byteCount*8 < coilCount {
		return nil, ErrInvalidPacket
	}
	values := make([]bool, coilCount)
	for i := uint16(0); i < coilCount; i++ {
		values[i] = pdu.Data[5+i/8]&(1<<uint(i%8)) != 0
	}
	return &WriteMultipleCoilsRequest{
		Offset: offset,
		Values: values,
	}, nil
}

type WriteMultipleCoilsResponse struct {
	Offset uint16
	Count  uint16
}

func (r *WriteMultipleCoilsResponse) Bytes() []byte {
	return []byte{
		byte(r.Offset >> 8),
		byte(r.Offset),
		byte(r.Count >> 8),
		byte(r.Count),
	}
}

type WriteMultipleRegistersRequest struct {
	Offset uint16
	Values []uint16
}

func NewWriteMultipleRegistersRequest(adu ApplicationDataUnit) (*WriteMultipleRegistersRequest, error) {
	pdu := adu.PDU()
	if len(pdu.Data) < 5 {
		return nil, ErrInvalidPacket
	}
	offset := uint16(pdu.Data[0])<<8 | uint16(pdu.Data[1])
	registerCount := uint16(pdu.Data[2])<<8 | uint16(pdu.Data[3])
	byteCount := uint16(pdu.Data[4])
	if uint16(len(pdu.Data)) != 2+2+byteCount+1 {
		return nil, ErrInvalidPacket
	}
	if byteCount != registerCount*2 {
		return nil, ErrInvalidPacket
	}
	values := make([]uint16, registerCount)
	for i := uint16(0); i < registerCount; i++ {
		values[i] = uint16(pdu.Data[5+i*2])<<8 | uint16(pdu.Data[6+i*2])
	}
	return &WriteMultipleRegistersRequest{
		Offset: offset,
		Values: values,
	}, nil
}

type WriteMultipleRegistersResponse struct {
	Offset uint16
	Count  uint16
}

func (r *WriteMultipleRegistersResponse) Bytes() []byte {
	return []byte{
		byte(r.Offset >> 8),
		byte(r.Offset),
		byte(r.Count >> 8),
		byte(r.Count),
	}
}
