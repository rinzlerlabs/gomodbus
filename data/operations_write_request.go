package data

import "go.uber.org/zap/zapcore"

type ModbusWriteSingleRequest[T bool | uint16] interface {
	ModbusOperation
	Offset() uint16
	Value() T
}

type ModbusWriteArrayRequest[T []bool | []uint16] interface {
	CountableOperation
	ModbusOperation
	Offset() uint16
	Values() T
}

func NewWriteSingleCoilRequest(offset uint16, value bool) *WriteSingleCoilRequest {
	return &WriteSingleCoilRequest{
		offset: offset,
		value:  value,
	}
}

type WriteSingleCoilRequest struct {
	ModbusWriteSingleRequest[bool]
	offset uint16
	value  bool
}

func (r WriteSingleCoilRequest) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddBool("Value", r.value)
	encoder.AddUint16("Offset", r.offset)
	return nil
}

func (r WriteSingleCoilRequest) Offset() uint16 {
	return r.offset
}

func (r WriteSingleCoilRequest) Value() bool {
	return r.value
}

func NewWriteSingleRegisterRequest(offset, value uint16) *WriteSingleRegisterRequest {
	return &WriteSingleRegisterRequest{
		offset: offset,
		value:  value,
	}
}

type WriteSingleRegisterRequest struct {
	ModbusWriteSingleRequest[uint16]
	offset uint16
	value  uint16
}

func (r WriteSingleRegisterRequest) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddUint16("Value", r.value)
	encoder.AddUint16("Offset", r.offset)
	return nil
}

func (r WriteSingleRegisterRequest) Offset() uint16 {
	return r.offset
}

func (r WriteSingleRegisterRequest) Value() uint16 {
	return r.value
}

func NewWriteMultipleCoilsRequest(offset uint16, values []bool) *WriteMultipleCoilsRequest {
	return &WriteMultipleCoilsRequest{
		offset: offset,
		values: values,
	}
}

type WriteMultipleCoilsRequest struct {
	ModbusWriteArrayRequest[[]bool]
	offset uint16
	values []bool
}

func (r WriteMultipleCoilsRequest) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddArray("Values", zapcore.ArrayMarshalerFunc(func(enc zapcore.ArrayEncoder) error {
		for _, v := range r.values {
			enc.AppendBool(v)
		}
		return nil
	}))
	encoder.AddUint16("Offset", r.offset)
	return nil
}

func (r WriteMultipleCoilsRequest) Offset() uint16 {
	return r.offset
}

func (r WriteMultipleCoilsRequest) Values() []bool {
	return r.values
}

func (r WriteMultipleCoilsRequest) Count() int {
	return len(r.values)
}

func NewWriteMultipleRegistersRequest(offset uint16, values []uint16) *WriteMultipleRegistersRequest {
	return &WriteMultipleRegistersRequest{
		offset: offset,
		values: values,
	}
}

type WriteMultipleRegistersRequest struct {
	ModbusWriteArrayRequest[[]uint16]
	offset uint16
	values []uint16
}

func (r WriteMultipleRegistersRequest) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddArray("Values", zapcore.ArrayMarshalerFunc(func(enc zapcore.ArrayEncoder) error {
		for _, v := range r.values {
			enc.AppendUint16(v)
		}
		return nil
	}))
	encoder.AddUint16("Offset", r.offset)
	return nil
}

func (r WriteMultipleRegistersRequest) Offset() uint16 {
	return r.offset
}

func (r WriteMultipleRegistersRequest) Values() []uint16 {
	return r.values
}

func (r WriteMultipleRegistersRequest) Count() int {
	return len(r.values)
}
