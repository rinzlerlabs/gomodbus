package data

import (
	"github.com/rinzlerlabs/gomodbus/common"
	"go.uber.org/zap/zapcore"
)

type FunctionCode byte
type ExceptionCode byte

const (
	ReadCoils                   FunctionCode = 0x01
	ReadDiscreteInputs          FunctionCode = 0x02
	ReadHoldingRegisters        FunctionCode = 0x03
	ReadInputRegisters          FunctionCode = 0x04
	WriteSingleCoil             FunctionCode = 0x05
	WriteSingleRegister         FunctionCode = 0x06
	WriteMultipleCoils          FunctionCode = 0x0F
	WriteMultipleRegisters      FunctionCode = 0x10
	ReadCoilsError              FunctionCode = 0x81
	ReadDiscreteInputsError     FunctionCode = 0x82
	ReadHoldingRegistersError   FunctionCode = 0x83
	ReadInputRegistersError     FunctionCode = 0x84
	WriteSingleCoilError        FunctionCode = 0x85
	WriteSingleRegisterError    FunctionCode = 0x86
	WriteMultipleCoilsError     FunctionCode = 0x8F
	WriteMultipleRegistersError FunctionCode = 0x90

	IllegalFunction                    ExceptionCode = 0x01
	IllegalDataAddress                 ExceptionCode = 0x02
	IllegalDataValue                   ExceptionCode = 0x03
	ServerDeviceFailure                ExceptionCode = 0x04
	Acknowledge                        ExceptionCode = 0x05
	ServerDeviceBusy                   ExceptionCode = 0x06
	MemoryParityError                  ExceptionCode = 0x08
	GatewayPathUnavailable             ExceptionCode = 0x0A
	GatewayTargetDeviceFailedToRespond ExceptionCode = 0x0B
)

func (f FunctionCode) String() string {
	switch f {
	case ReadCoils:
		return "ReadCoils"
	case ReadDiscreteInputs:
		return "ReadDiscreteInputs"
	case ReadHoldingRegisters:
		return "ReadHoldingRegisters"
	case ReadInputRegisters:
		return "ReadInputRegisters"
	case WriteSingleCoil:
		return "WriteSingleCoil"
	case WriteSingleRegister:
		return "WriteSingleRegister"
	case WriteMultipleCoils:
		return "WriteMultipleCoils"
	case WriteMultipleRegisters:
		return "WriteMultipleRegisters"
	default:
		return "Unknown"
	}
}

func (f FunctionCode) IsException() bool {
	return f >= 0x80 && f <= 0x91
}

func (f ExceptionCode) String() string {
	switch f {
	case IllegalFunction:
		return "IllegalFunction"
	case IllegalDataAddress:
		return "IllegalDataAddress"
	case IllegalDataValue:
		return "IllegalDataValue"
	case ServerDeviceFailure:
		return "SlaveDeviceFailure"
	case Acknowledge:
		return "Acknowledge"
	case ServerDeviceBusy:
		return "SlaveDeviceBusy"
	case MemoryParityError:
		return "MemoryParityError"
	case GatewayPathUnavailable:
		return "GatewayPathUnavailable"
	case GatewayTargetDeviceFailedToRespond:
		return "GatewayTargetDeviceFailedToRespond"
	default:
		return "UnknownException"
	}
}

type ModbusOperation interface {
	zapcore.ObjectMarshaler
	Bytes() []byte
}

type CountableRequest interface {
	ValueCount() uint16
}

type ModbusReadOperation[T any] interface {
	ModbusOperation
	Value() T
}

func NewReadCoilsRequest(offset, count uint16) *ReadCoilsRequest {
	return &ReadCoilsRequest{
		Offset: offset,
		Count:  count,
	}
}

type ReadCoilsRequest struct {
	CountableRequest
	Offset uint16
	Count  uint16
}

func (r ReadCoilsRequest) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddUint16("Count", uint16(r.Count))
	encoder.AddUint16("Offset", r.Offset)
	return nil
}

func (r *ReadCoilsRequest) Bytes() []byte {
	return []byte{
		byte(r.Offset >> 8),
		byte(r.Offset),
		byte(r.Count >> 8),
		byte(r.Count),
	}
}

func (r *ReadCoilsRequest) ValueCount() uint16 {
	return r.Count
}

func NewReadCoilsResponse(values []bool) *ReadCoilsResponse {
	return &ReadCoilsResponse{
		Values: values,
	}
}

type ReadCoilsResponse struct {
	Values []bool
}

func (r ReadCoilsResponse) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddArray("Values", zapcore.ArrayMarshalerFunc(func(enc zapcore.ArrayEncoder) error {
		for _, v := range r.Values {
			enc.AppendBool(v)
		}
		return nil
	}))
	return nil
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

func (r ReadCoilsResponse) Value() []bool {
	return r.Values
}

func getReturnByteCount(values []bool) byte {
	if len(values)%8 == 0 {
		return byte(len(values) / 8)
	}
	return byte(len(values)/8 + 1)
}

func NewReadDiscreteInputsRequest(offset, count uint16) *ReadDiscreteInputsRequest {
	return &ReadDiscreteInputsRequest{
		Offset: offset,
		Count:  count,
	}
}

type ReadDiscreteInputsRequest struct {
	CountableRequest
	Offset uint16
	Count  uint16
}

func (r *ReadDiscreteInputsRequest) ValueCount() uint16 {
	return r.Count
}

func (r *ReadDiscreteInputsRequest) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddUint16("Count", uint16(r.Count))
	encoder.AddUint16("Offset", r.Offset)
	return nil
}

func (r *ReadDiscreteInputsRequest) Bytes() []byte {
	return []byte{
		byte(r.Offset >> 8),
		byte(r.Offset),
		byte(r.Count >> 8),
		byte(r.Count),
	}
}

func NewReadDiscreteInputsResponse(values []bool) *ReadDiscreteInputsResponse {
	return &ReadDiscreteInputsResponse{
		Values: values,
	}
}

type ReadDiscreteInputsResponse struct {
	Values []bool
}

func (r ReadDiscreteInputsResponse) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddArray("Values", zapcore.ArrayMarshalerFunc(func(enc zapcore.ArrayEncoder) error {
		for _, v := range r.Values {
			enc.AppendBool(v)
		}
		return nil
	}))
	return nil
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

func NewReadHoldingRegistersRequest(offset, count uint16) *ReadHoldingRegistersRequest {
	return &ReadHoldingRegistersRequest{
		Offset: offset,
		Count:  count,
	}
}

type ReadHoldingRegistersRequest struct {
	CountableRequest
	Offset uint16
	Count  uint16
}

func (r *ReadHoldingRegistersRequest) ValueCount() uint16 {
	return r.Count
}

func (r ReadHoldingRegistersRequest) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddUint16("Count", uint16(r.Count))
	encoder.AddUint16("Offset", r.Offset)
	return nil
}

func (r *ReadHoldingRegistersRequest) Bytes() []byte {
	return []byte{
		byte(r.Offset >> 8),
		byte(r.Offset),
		byte(r.Count >> 8),
		byte(r.Count),
	}
}

func NewReadHoldingRegistersResponse(values []uint16) *ReadHoldingRegistersResponse {
	return &ReadHoldingRegistersResponse{
		Values: values,
	}
}

type ReadHoldingRegistersResponse struct {
	Values []uint16
}

func (r ReadHoldingRegistersResponse) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddArray("Values", zapcore.ArrayMarshalerFunc(func(enc zapcore.ArrayEncoder) error {
		for _, v := range r.Values {
			enc.AppendUint16(v)
		}
		return nil
	}))
	return nil
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

func NewReadInputRegistersRequest(offset, count uint16) *ReadInputRegistersRequest {
	return &ReadInputRegistersRequest{
		Offset: offset,
		Count:  count,
	}
}

type ReadInputRegistersRequest struct {
	CountableRequest
	Offset uint16
	Count  uint16
}

func (r *ReadInputRegistersRequest) ValueCount() uint16 {
	return r.Count
}

func (r *ReadInputRegistersRequest) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddUint16("Count", uint16(r.Count))
	encoder.AddUint16("Offset", r.Offset)
	return nil
}

func (r *ReadInputRegistersRequest) Bytes() []byte {
	return []byte{
		byte(r.Offset >> 8),
		byte(r.Offset),
		byte(r.Count >> 8),
		byte(r.Count),
	}
}

func NewReadInputRegistersResponse(values []uint16) *ReadInputRegistersResponse {
	return &ReadInputRegistersResponse{
		Values: values,
	}
}

type ReadInputRegistersResponse struct {
	Values []uint16
}

func (r ReadInputRegistersResponse) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddArray("Values", zapcore.ArrayMarshalerFunc(func(enc zapcore.ArrayEncoder) error {
		for _, v := range r.Values {
			enc.AppendUint16(v)
		}
		return nil
	}))
	return nil
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

func NewWriteSingleCoilRequest(offset uint16, value bool) *WriteSingleCoilRequest {
	return &WriteSingleCoilRequest{
		Offset: offset,
		Value:  value,
	}
}

type WriteSingleCoilRequest struct {
	Offset uint16
	Value  bool
}

func (r WriteSingleCoilRequest) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddBool("Value", r.Value)
	encoder.AddUint16("Offset", r.Offset)
	return nil
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

func NewWriteSingleCoilResponse(offset uint16, value bool) *WriteSingleCoilResponse {
	return &WriteSingleCoilResponse{
		Offset: offset,
		Value:  value,
	}
}

type WriteSingleCoilResponse struct {
	Offset uint16
	Value  bool
}

func (r WriteSingleCoilResponse) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddBool("Value", r.Value)
	encoder.AddUint16("Offset", r.Offset)
	return nil
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

func NewWriteSingleRegisterRequest(offset, value uint16) *WriteSingleRegisterRequest {
	return &WriteSingleRegisterRequest{
		Offset: offset,
		Value:  value,
	}
}

type WriteSingleRegisterRequest struct {
	Offset uint16
	Value  uint16
}

func (r WriteSingleRegisterRequest) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddUint16("Value", r.Value)
	encoder.AddUint16("Offset", r.Offset)
	return nil
}

func (r *WriteSingleRegisterRequest) Bytes() []byte {
	return []byte{
		byte(r.Offset >> 8),
		byte(r.Offset),
		byte(r.Value >> 8),
		byte(r.Value),
	}
}

func NewWriteSingleRegisterResponse(offset, value uint16) *WriteSingleRegisterResponse {
	return &WriteSingleRegisterResponse{
		Offset: offset,
		Value:  value,
	}
}

type WriteSingleRegisterResponse struct {
	Offset uint16
	Value  uint16
}

func (r WriteSingleRegisterResponse) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddUint16("Value", r.Value)
	encoder.AddUint16("Offset", r.Offset)
	return nil
}

func (r *WriteSingleRegisterResponse) Bytes() []byte {
	return []byte{
		byte(r.Offset >> 8),
		byte(r.Offset),
		byte(r.Value >> 8),
		byte(r.Value),
	}
}

func NewWriteMultipleCoilsRequest(offset uint16, values []bool) *WriteMultipleCoilsRequest {
	return &WriteMultipleCoilsRequest{
		Offset: offset,
		Values: values,
	}
}

type WriteMultipleCoilsRequest struct {
	Offset uint16
	Values []bool
}

func (r *WriteMultipleCoilsRequest) ValueCount() int {
	return len(r.Values)
}

func (r WriteMultipleCoilsRequest) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddArray("Values", zapcore.ArrayMarshalerFunc(func(enc zapcore.ArrayEncoder) error {
		for _, v := range r.Values {
			enc.AppendBool(v)
		}
		return nil
	}))
	encoder.AddUint16("Offset", r.Offset)
	return nil
}

func (r *WriteMultipleCoilsRequest) Bytes() []byte {
	valueCount := len(r.Values)
	byteCount := getReturnByteCount(r.Values)
	data := make([]byte, 5+byteCount)
	data[0] = byte(r.Offset >> 8)
	data[1] = byte(r.Offset)
	data[2] = byte(valueCount >> 8)
	data[3] = byte(valueCount)
	data[4] = byte(byteCount)
	for i, v := range r.Values {
		if v {
			data[5+i/8] |= 1 << uint(i%8)
		}
	}
	return data
}

func NewWriteMultipleCoilsResponse(offset, count uint16) *WriteMultipleCoilsResponse {
	return &WriteMultipleCoilsResponse{
		Offset: offset,
		Count:  count,
	}
}

type WriteMultipleCoilsResponse struct {
	Offset uint16
	Count  uint16
}

func (r WriteMultipleCoilsResponse) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddUint16("Count", r.Count)
	encoder.AddUint16("Offset", r.Offset)
	return nil
}

func (r *WriteMultipleCoilsResponse) Bytes() []byte {
	return []byte{
		byte(r.Offset >> 8),
		byte(r.Offset),
		byte(r.Count >> 8),
		byte(r.Count),
	}
}

func NewWriteMultipleRegistersRequest(offset uint16, values []uint16) *WriteMultipleRegistersRequest {
	return &WriteMultipleRegistersRequest{
		Offset: offset,
		Values: values,
	}
}

type WriteMultipleRegistersRequest struct {
	Offset uint16
	Values []uint16
}

func (r *WriteMultipleRegistersRequest) ValueCount() int {
	return len(r.Values)
}

func (r WriteMultipleRegistersRequest) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddArray("Values", zapcore.ArrayMarshalerFunc(func(enc zapcore.ArrayEncoder) error {
		for _, v := range r.Values {
			enc.AppendUint16(v)
		}
		return nil
	}))
	encoder.AddUint16("Offset", r.Offset)
	return nil
}

func (r *WriteMultipleRegistersRequest) Bytes() []byte {
	valueCount := len(r.Values)
	byteCount := 2 * valueCount
	data := make([]byte, 5+byteCount)
	data[0] = byte(r.Offset >> 8)
	data[1] = byte(r.Offset)
	data[2] = byte(valueCount >> 8)
	data[3] = byte(valueCount)
	data[4] = byte(byteCount)
	for i, v := range r.Values {
		data[5+i*2] = byte(v >> 8)
		data[6+i*2] = byte(v)
	}
	return data
}

func NewWriteMultipleRegistersResponse(offset, count uint16) *WriteMultipleRegistersResponse {
	return &WriteMultipleRegistersResponse{
		Offset: offset,
		Count:  count,
	}
}

type WriteMultipleRegistersResponse struct {
	Offset uint16
	Count  uint16
}

func (r WriteMultipleRegistersResponse) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddUint16("Count", r.Count)
	encoder.AddUint16("Offset", r.Offset)
	return nil
}

func (r *WriteMultipleRegistersResponse) Bytes() []byte {
	return []byte{
		byte(r.Offset >> 8),
		byte(r.Offset),
		byte(r.Count >> 8),
		byte(r.Count),
	}
}

func NewModbusOperationException(requestFunction FunctionCode, code ExceptionCode) *ModbusOperationException {
	return &ModbusOperationException{
		ExceptionCode: code,
	}
}

type ModbusOperationException struct {
	ExceptionCode ExceptionCode
}

func (e ModbusOperationException) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("ExceptionCode", e.ExceptionCode.String())
	return nil
}

func (e *ModbusOperationException) Bytes() []byte {
	return []byte{byte(e.ExceptionCode)}
}

func (e *ModbusOperationException) Error() error {
	switch e.ExceptionCode {
	case IllegalFunction:
		return common.ErrIllegalFunction
	case IllegalDataAddress:
		return common.ErrIllegalDataAddress
	case IllegalDataValue:
		return common.ErrIllegalDataValue
	case ServerDeviceFailure:
		return common.ErrServerDeviceFailure
	case Acknowledge:
		return common.ErrAcknowledge
	case ServerDeviceBusy:
		return common.ErrServerDeviceBusy
	case MemoryParityError:
		return common.ErrMemoryParityError
	case GatewayPathUnavailable:
		return common.ErrGatewayPathUnavailable
	case GatewayTargetDeviceFailedToRespond:
		return common.ErrGatewayTargetDeviceFailedToRespond
	default:
		return common.ErrUnknownException
	}
}
