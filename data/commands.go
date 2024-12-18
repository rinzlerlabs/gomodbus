package data

type FunctionCode byte

const (
	ReadCoils              FunctionCode = 0x01
	ReadDiscreteInputs     FunctionCode = 0x02
	ReadHoldingRegisters   FunctionCode = 0x03
	ReadInputRegisters     FunctionCode = 0x04
	WriteSingleCoil        FunctionCode = 0x05
	WriteSingleRegister    FunctionCode = 0x06
	WriteMultipleCoils     FunctionCode = 0x0F
	WriteMultipleRegisters FunctionCode = 0x10
)

type ModbusRequest interface {
	Bytes() []byte
}

type ModbusResponse interface {
	Bytes() []byte
}

type ReadCoilsRequest struct {
	ModbusRequest
	Offset uint16
	Count  uint16
}

func (r *ReadCoilsRequest) Bytes() []byte {
	return []byte{
		byte(r.Offset >> 8),
		byte(r.Offset),
		byte(r.Count >> 8),
		byte(r.Count),
	}
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

func (r *ReadDiscreteInputsRequest) Bytes() []byte {
	return []byte{
		byte(r.Offset >> 8),
		byte(r.Offset),
		byte(r.Count >> 8),
		byte(r.Count),
	}
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

func (r *ReadHoldingRegistersRequest) Bytes() []byte {
	return []byte{
		byte(r.Offset >> 8),
		byte(r.Offset),
		byte(r.Count >> 8),
		byte(r.Count),
	}
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

func (r *ReadInputRegistersRequest) Bytes() []byte {
	return []byte{
		byte(r.Offset >> 8),
		byte(r.Offset),
		byte(r.Count >> 8),
		byte(r.Count),
	}
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

func (r *WriteSingleCoilRequest) Bytes() []byte {
	val := byte(0x00)
	if r.Value {
		val = 0xFF
	}
	return []byte{
		byte(r.Offset >> 8),
		byte(r.Offset),
		val,
		0x00,
	}
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
		val,
		0x00,
	}
}

type WriteSingleRegisterRequest struct {
	Offset uint16
	Value  uint16
}

func (r *WriteSingleRegisterRequest) Bytes() []byte {
	return []byte{
		byte(r.Offset >> 8),
		byte(r.Offset),
		byte(r.Value >> 8),
		byte(r.Value),
	}
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

func (r *WriteMultipleCoilsRequest) Bytes() []byte {
	byteCount := getReturnByteCount(r.Values)
	data := make([]byte, 5+byteCount)
	data[0] = byte(r.Offset >> 8)
	data[1] = byte(r.Offset)
	data[2] = byte(len(r.Values) >> 8)
	data[3] = byte(len(r.Values))
	data[4] = byte(byteCount)
	for i, v := range r.Values {
		if v {
			data[5+i/8] |= 1 << uint(i%8)
		}
	}
	return data
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

func (r *WriteMultipleRegistersRequest) Bytes() []byte {
	byteCount := 2 * len(r.Values)
	data := make([]byte, 5+byteCount)
	data[0] = byte(r.Offset >> 8)
	data[1] = byte(r.Offset)
	data[2] = byte(len(r.Values) >> 8)
	data[3] = byte(len(r.Values))
	data[4] = byte(byteCount)
	for i, v := range r.Values {
		data[5+i*2] = byte(v >> 8)
		data[6+i*2] = byte(v)
	}
	return data
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
