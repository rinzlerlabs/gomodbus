package data

import "go.uber.org/zap/zapcore"

type ModbusReadResponse[T []bool | []uint16] interface {
	ModbusOperation
	CountableOperation
	Values() T
}

func NewReadCoilsResponse(values []bool) *ReadCoilsResponse {
	return &ReadCoilsResponse{
		values: values,
	}
}

type ReadCoilsResponse struct {
	ModbusReadResponse[[]bool]
	values []bool
}

func (r ReadCoilsResponse) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddArray("Values", zapcore.ArrayMarshalerFunc(func(enc zapcore.ArrayEncoder) error {
		for _, v := range r.values {
			enc.AppendBool(v)
		}
		return nil
	}))
	return nil
}

func (r ReadCoilsResponse) Values() []bool {
	return r.values
}

func (r ReadCoilsResponse) Count() int {
	return len(r.values)
}

func NewReadDiscreteInputsResponse(values []bool) *ReadDiscreteInputsResponse {
	return &ReadDiscreteInputsResponse{
		values: values,
	}
}

type ReadDiscreteInputsResponse struct {
	ModbusReadResponse[[]bool]
	values []bool
}

func (r ReadDiscreteInputsResponse) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddArray("Values", zapcore.ArrayMarshalerFunc(func(enc zapcore.ArrayEncoder) error {
		for _, v := range r.values {
			enc.AppendBool(v)
		}
		return nil
	}))
	return nil
}

func (r ReadDiscreteInputsResponse) Values() []bool {
	return r.values
}

func (r ReadDiscreteInputsResponse) Count() int {
	return len(r.values)
}

func NewReadHoldingRegistersResponse(values []uint16) *ReadHoldingRegistersResponse {
	return &ReadHoldingRegistersResponse{
		values: values,
	}
}

type ReadHoldingRegistersResponse struct {
	ModbusReadResponse[[]uint16]
	values []uint16
}

func (r ReadHoldingRegistersResponse) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddArray("Values", zapcore.ArrayMarshalerFunc(func(enc zapcore.ArrayEncoder) error {
		for _, v := range r.values {
			enc.AppendUint16(v)
		}
		return nil
	}))
	return nil
}

func (r ReadHoldingRegistersResponse) Values() []uint16 {
	return r.values
}

func (r ReadHoldingRegistersResponse) Count() int {
	return len(r.values)
}

func NewReadInputRegistersResponse(values []uint16) *ReadInputRegistersResponse {
	return &ReadInputRegistersResponse{
		values: values,
	}
}

type ReadInputRegistersResponse struct {
	ModbusReadResponse[[]uint16]
	values []uint16
}

func (r ReadInputRegistersResponse) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddArray("Values", zapcore.ArrayMarshalerFunc(func(enc zapcore.ArrayEncoder) error {
		for _, v := range r.values {
			enc.AppendUint16(v)
		}
		return nil
	}))
	return nil
}

func (r ReadInputRegistersResponse) Values() []uint16 {
	return r.values
}

func (r ReadInputRegistersResponse) Count() int {
	return len(r.values)
}
