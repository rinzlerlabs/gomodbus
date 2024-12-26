package data

import "go.uber.org/zap/zapcore"

type ModbusWriteSingleResponse[T bool | uint16] interface {
	ModbusOperation
	Offset() uint16
	Value() T
}

type ModbusWriteArrayResponse[T []bool | []uint16] interface {
	CountableOperation
	ModbusOperation
	Offset() uint16
	Value() T
}

func NewWriteSingleCoilResponse(offset uint16, value bool) *WriteSingleCoilResponse {
	return &WriteSingleCoilResponse{
		offset: offset,
		value:  value,
	}
}

type WriteSingleCoilResponse struct {
	ModbusWriteSingleResponse[bool]
	offset uint16
	value  bool
}

func (r WriteSingleCoilResponse) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddBool("Value", r.value)
	encoder.AddUint16("Offset", r.offset)
	return nil
}

func (r WriteSingleCoilResponse) Offset() uint16 {
	return r.offset
}

func (r WriteSingleCoilResponse) Value() bool {
	return r.value
}

func NewWriteSingleRegisterResponse(offset, value uint16) *WriteSingleRegisterResponse {
	return &WriteSingleRegisterResponse{
		offset: offset,
		value:  value,
	}
}

type WriteSingleRegisterResponse struct {
	ModbusWriteSingleResponse[uint16]
	offset uint16
	value  uint16
}

func (r WriteSingleRegisterResponse) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddUint16("Value", r.value)
	encoder.AddUint16("Offset", r.offset)
	return nil
}

func (r WriteSingleRegisterResponse) Offset() uint16 {
	return r.offset
}

func (r WriteSingleRegisterResponse) Value() uint16 {
	return r.value
}

func NewWriteMultipleCoilsResponse(offset, count uint16) *WriteMultipleCoilsResponse {
	return &WriteMultipleCoilsResponse{
		offset: offset,
		count:  count,
	}
}

type WriteMultipleCoilsResponse struct {
	ModbusWriteArrayResponse[[]bool]
	offset uint16
	count  uint16
}

func (r WriteMultipleCoilsResponse) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddUint16("Count", r.count)
	encoder.AddUint16("Offset", r.offset)
	return nil
}

func (r WriteMultipleCoilsResponse) Offset() uint16 {
	return r.offset
}

func (r WriteMultipleCoilsResponse) Count() int {
	return int(r.count)
}

func NewWriteMultipleRegistersResponse(offset, count uint16) *WriteMultipleRegistersResponse {
	return &WriteMultipleRegistersResponse{
		offset: offset,
		count:  count,
	}
}

type WriteMultipleRegistersResponse struct {
	ModbusWriteArrayResponse[[]uint16]
	offset uint16
	count  uint16
}

func (r WriteMultipleRegistersResponse) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddUint16("Count", r.count)
	encoder.AddUint16("Offset", r.offset)
	return nil
}

func (r WriteMultipleRegistersResponse) Offset() uint16 {
	return r.offset
}

func (r WriteMultipleRegistersResponse) Count() int {
	return int(r.count)
}
