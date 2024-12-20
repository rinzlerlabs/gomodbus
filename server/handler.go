package server

import (
	"github.com/rinzlerlabs/gomodbus/common"
	"github.com/rinzlerlabs/gomodbus/data"
	"github.com/rinzlerlabs/gomodbus/transport"
	"go.uber.org/zap"
)

// RequestHandler is the interface that wraps the basic Modbus functions.
// TODO: Merge single and multiple requests into one.
type RequestHandler interface {
	Handle(op transport.ModbusTransaction) error
	ReadCoils(request data.ModbusOperation) (response data.ModbusReadOperation[[]bool], err error)
	ReadDiscreteInputs(request data.ModbusOperation) (response *data.ReadDiscreteInputsResponse, err error)
	ReadHoldingRegisters(request data.ModbusOperation) (response *data.ReadHoldingRegistersResponse, err error)
	ReadInputRegisters(request data.ModbusOperation) (response *data.ReadInputRegistersResponse, err error)
	WriteSingleCoil(equest data.ModbusOperation) (response *data.WriteSingleCoilResponse, err error)
	WriteSingleRegister(request data.ModbusOperation) (response *data.WriteSingleRegisterResponse, err error)
	WriteMultipleCoils(request data.ModbusOperation) (response *data.WriteMultipleCoilsResponse, err error)
	WriteMultipleRegisters(request data.ModbusOperation) (response *data.WriteMultipleRegistersResponse, err error)
}

type DefaultHandler struct {
	logger           *zap.Logger
	Coils            []bool
	DiscreteInputs   []bool
	HoldingRegisters []uint16
	InputRegisters   []uint16
}

func NewDefaultHandler(logger *zap.Logger, coilCount, discreteInputCount, holdingRegisterCount, inputRegisterCount uint16) RequestHandler {
	if logger == nil {
		logger = zap.NewNop()
	}
	if coilCount == 0 || discreteInputCount == 0 || holdingRegisterCount == 0 || inputRegisterCount == 0 {
		logger.Warn("Invalid count, using default values")
		coilCount = 65535
		discreteInputCount = 65535
		holdingRegisterCount = 65535
		inputRegisterCount = 65535
	}
	return &DefaultHandler{
		logger:           logger,
		Coils:            make([]bool, coilCount),
		DiscreteInputs:   make([]bool, discreteInputCount),
		HoldingRegisters: make([]uint16, holdingRegisterCount),
		InputRegisters:   make([]uint16, inputRegisterCount),
	}
}

func (h *DefaultHandler) Handle(txn transport.ModbusTransaction) error {
	h.logger.Info("Received request", zap.Any("Operation", txn.Frame().PDU().Operation()))
	var result data.ModbusOperation
	var err error
	switch txn.Frame().PDU().FunctionCode() {
	case data.ReadCoils:
		// Read Coils
		result, err = h.ReadCoils(txn.Frame().PDU().Operation())
	case data.ReadDiscreteInputs:
		// Read Discrete Inputs
		result, err = h.ReadDiscreteInputs(txn.Frame().PDU().Operation())
	case data.ReadHoldingRegisters:
		// Read Holding Registers
		result, err = h.ReadHoldingRegisters(txn.Frame().PDU().Operation())
	case data.ReadInputRegisters:
		// Read Input Registers
		result, err = h.ReadInputRegisters(txn.Frame().PDU().Operation())
	case data.WriteSingleCoil:
		// Write Single Coil
		result, err = h.WriteSingleCoil(txn.Frame().PDU().Operation())
	case data.WriteSingleRegister:
		// Write Single Register
		result, err = h.WriteSingleRegister(txn.Frame().PDU().Operation())
	case data.WriteMultipleCoils:
		// Write Multiple Coils
		result, err = h.WriteMultipleCoils(txn.Frame().PDU().Operation())
	case data.WriteMultipleRegisters:
		// Write Multiple Registers
		result, err = h.WriteMultipleRegisters(txn.Frame().PDU().Operation())
	default:
		h.logger.Debug("Received packet with unknown function code, discarding packet", zap.Any("packet", txn))
		return common.ErrUnknownFunctionCode
	}
	if err != nil {
		h.logger.Error("Failed to handle request", zap.Error(err))
		return err
	}
	h.logger.Info("Request handled successfully", zap.Any("response", result))
	return txn.Write(transport.NewProtocolDataUnit(txn.Frame().PDU().FunctionCode(), result))
}

func getRange(offset, count uint16) (uint16, uint16) {
	start := offset + 1
	end := 1 + offset + count
	return start, end
}

func (h *DefaultHandler) ReadCoils(operation data.ModbusOperation) (response data.ModbusReadOperation[[]bool], err error) {
	op := operation.(*data.ReadCoilsRequest)
	h.logger.Debug("ReadCoils", zap.Uint16("Offset", op.Offset), zap.Uint16("Count", op.Count))
	start, end := getRange(op.Offset, op.Count)
	results := h.Coils[start:end]
	return data.NewReadCoilsResponse(results), nil
}

func (h *DefaultHandler) ReadDiscreteInputs(operation data.ModbusOperation) (response *data.ReadDiscreteInputsResponse, err error) {
	op := operation.(*data.ReadDiscreteInputsRequest)
	h.logger.Debug("ReadDiscreteInputs", zap.Uint16("Offset", op.Offset), zap.Uint16("Count", op.Count))
	start, end := getRange(op.Offset, op.Count)
	results := h.DiscreteInputs[start:end]
	return data.NewReadDiscreteInputsResponse(results), nil
}

func (h *DefaultHandler) ReadHoldingRegisters(operation data.ModbusOperation) (response *data.ReadHoldingRegistersResponse, err error) {
	op := operation.(*data.ReadHoldingRegistersRequest)
	h.logger.Debug("ReadHoldingRegisters", zap.Uint16("Offset", op.Offset), zap.Uint16("Count", op.Count))
	start, end := getRange(op.Offset, op.Count)
	results := h.HoldingRegisters[start:end]
	return data.NewReadHoldingRegistersResponse(results), nil
}

func (h *DefaultHandler) ReadInputRegisters(operation data.ModbusOperation) (response *data.ReadInputRegistersResponse, err error) {
	op := operation.(*data.ReadInputRegistersRequest)
	h.logger.Debug("ReadInputRegisters", zap.Uint16("Offset", op.Offset), zap.Uint16("Count", op.Count))
	start, end := getRange(op.Offset, op.Count)
	results := h.InputRegisters[start:end]
	return data.NewReadInputRegistersResponse(results), nil
}

func (h *DefaultHandler) WriteSingleCoil(operation data.ModbusOperation) (response *data.WriteSingleCoilResponse, err error) {
	op := operation.(*data.WriteSingleCoilRequest)
	h.logger.Debug("WriteSingleCoil", zap.Uint16("Offset", op.Offset), zap.Bool("Value", op.Value))
	h.Coils[op.Offset+1] = op.Value
	return data.NewWriteSingleCoilResponse(op.Offset, op.Value), nil
}

func (h *DefaultHandler) WriteSingleRegister(operation data.ModbusOperation) (response *data.WriteSingleRegisterResponse, err error) {
	op := operation.(*data.WriteSingleRegisterRequest)
	h.logger.Debug("WriteSingleRegister", zap.Uint16("Offset", op.Offset), zap.Uint16("Value", op.Value))
	h.HoldingRegisters[op.Offset+1] = op.Value
	return data.NewWriteSingleRegisterResponse(op.Offset, op.Value), nil
}

func (h *DefaultHandler) WriteMultipleCoils(operation data.ModbusOperation) (response *data.WriteMultipleCoilsResponse, err error) {
	op := operation.(*data.WriteMultipleCoilsRequest)
	h.logger.Debug("WriteMultipleCoils", zap.Uint16("Offset", op.Offset), zap.Bools("Values", op.Values))
	start, _ := getRange(op.Offset, uint16(len(op.Values)))
	for i, v := range op.Values {
		h.Coils[start+uint16(i)] = v
	}
	return data.NewWriteMultipleCoilsResponse(op.Offset, uint16(len(op.Values))), nil
}

func (h *DefaultHandler) WriteMultipleRegisters(operation data.ModbusOperation) (response *data.WriteMultipleRegistersResponse, err error) {
	op := operation.(*data.WriteMultipleRegistersRequest)
	h.logger.Debug("WriteMultipleRegisters", zap.Uint16("Offset", op.Offset), zap.Uint16s("Values", op.Values))
	start, _ := getRange(op.Offset, uint16(len(op.Values)))
	for i, v := range op.Values {
		h.HoldingRegisters[start+uint16(i)] = v
	}
	return data.NewWriteMultipleRegistersResponse(op.Offset, uint16(len(op.Values))), nil
}
