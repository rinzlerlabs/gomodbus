package data

import "go.uber.org/zap/zapcore"

type ModbusReadRequest interface {
	ModbusOperation
	CountableOperation
	Offset() uint16
}

func NewReadCoilsRequest(offset, count uint16) *ReadCoilsRequest {
	return &ReadCoilsRequest{
		offset: offset,
		count:  count,
	}
}

type ReadCoilsRequest struct {
	ModbusReadRequest
	offset uint16
	count  uint16
}

func (r ReadCoilsRequest) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddUint16("Count", uint16(r.count))
	encoder.AddUint16("Offset", r.offset)
	return nil
}

func (r ReadCoilsRequest) Offset() uint16 {
	return r.offset
}

func (r ReadCoilsRequest) Count() int {
	return int(r.count)
}

func NewReadDiscreteInputsRequest(offset, count uint16) *ReadDiscreteInputsRequest {
	return &ReadDiscreteInputsRequest{
		offset: offset,
		count:  count,
	}
}

type ReadDiscreteInputsRequest struct {
	ModbusReadRequest
	offset uint16
	count  uint16
}

func (r ReadDiscreteInputsRequest) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddUint16("Count", uint16(r.count))
	encoder.AddUint16("Offset", r.offset)
	return nil
}

func (r ReadDiscreteInputsRequest) Offset() uint16 {
	return r.offset
}

func (r ReadDiscreteInputsRequest) Count() int {
	return int(r.count)
}

func NewReadHoldingRegistersRequest(offset, count uint16) *ReadHoldingRegistersRequest {
	return &ReadHoldingRegistersRequest{
		offset: offset,
		count:  count,
	}
}

type ReadHoldingRegistersRequest struct {
	ModbusReadRequest
	offset uint16
	count  uint16
}

func (r ReadHoldingRegistersRequest) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddUint16("Count", uint16(r.count))
	encoder.AddUint16("Offset", r.offset)
	return nil
}

func (r ReadHoldingRegistersRequest) Offset() uint16 {
	return r.offset
}

func (r ReadHoldingRegistersRequest) Count() int {
	return int(r.count)
}

func NewReadInputRegistersRequest(offset, count uint16) *ReadInputRegistersRequest {
	return &ReadInputRegistersRequest{
		offset: offset,
		count:  count,
	}
}

type ReadInputRegistersRequest struct {
	ModbusReadRequest
	offset uint16
	count  uint16
}

func (r ReadInputRegistersRequest) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddUint16("Count", uint16(r.count))
	encoder.AddUint16("Offset", r.offset)
	return nil
}

func (r ReadInputRegistersRequest) Offset() uint16 {
	return r.offset
}

func (r ReadInputRegistersRequest) Count() int {
	return int(r.count)
}
