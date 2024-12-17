package gomodbus

import (
	"context"

	"go.uber.org/zap"
)

// RequestHandler is the interface that wraps the basic Modbus functions.
// TODO: Merge single and multiple requests into one.
type RequestHandler interface {
	ReadCoils(ctx context.Context, request *ReadCoilsRequest) (response *ReadCoilsResponse, err error)
	ReadDiscreteInputs(ctx context.Context, request *ReadDiscreteInputsRequest) (response *ReadDiscreteInputsResponse, err error)
	ReadHoldingRegisters(ctx context.Context, request *ReadHoldingRegistersRequest) (response *ReadHoldingRegistersResponse, err error)
	ReadInputRegisters(ctx context.Context, request *ReadInputRegistersRequest) (response *ReadInputRegistersResponse, err error)
	WriteSingleCoil(rctx context.Context, equest *WriteSingleCoilRequest) (response *WriteSingleCoilResponse, err error)
	WriteSingleRegister(ctx context.Context, request *WriteSingleRegisterRequest) (response *WriteSingleRegisterResponse, err error)
	WriteMultipleCoils(ctx context.Context, request *WriteMultipleCoilsRequest) (response *WriteMultipleCoilsResponse, err error)
	WriteMultipleRegisters(ctx context.Context, request *WriteMultipleRegistersRequest) (response *WriteMultipleRegistersResponse, err error)
}

type DefaultHandler struct {
	logger           *zap.Logger
	Coils            []bool
	DiscreteInputs   []bool
	HoldingRegisters []uint16
	InputRegisters   []uint16
}

func NewDefaultHandler(logger *zap.Logger, coilCount, discreteInputCount, holdingRegisterCount, inputRegisterCount uint16) *DefaultHandler {
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

func (h *DefaultHandler) ReadCoils(ctx context.Context, request *ReadCoilsRequest) (response *ReadCoilsResponse, err error) {
	h.logger.Debug("ReadCoils", zap.Uint16("Offset", request.Offset), zap.Uint16("Count", request.Count))
	start := 1 + request.Offset
	end := 1 + request.Offset + request.Count
	results := h.Coils[start:end]
	return &ReadCoilsResponse{Values: results}, nil
}

func (h *DefaultHandler) ReadDiscreteInputs(ctx context.Context, request *ReadDiscreteInputsRequest) (response *ReadDiscreteInputsResponse, err error) {
	h.logger.Debug("ReadDiscreteInputs", zap.Uint16("Offset", request.Offset), zap.Uint16("Count", request.Count))
	start := 1 + request.Offset
	end := 1 + request.Offset + request.Count
	results := h.DiscreteInputs[start:end]
	return &ReadDiscreteInputsResponse{Values: results}, nil
}

func (h *DefaultHandler) ReadHoldingRegisters(ctx context.Context, request *ReadHoldingRegistersRequest) (response *ReadHoldingRegistersResponse, err error) {
	h.logger.Debug("ReadHoldingRegisters", zap.Uint16("Offset", request.Offset), zap.Uint16("Count", request.Count))
	start := 1 + request.Offset
	end := 1 + request.Offset + request.Count
	results := h.HoldingRegisters[start:end]
	return &ReadHoldingRegistersResponse{Values: results}, nil
}

func (h *DefaultHandler) ReadInputRegisters(ctx context.Context, request *ReadInputRegistersRequest) (response *ReadInputRegistersResponse, err error) {
	h.logger.Debug("ReadInputRegisters", zap.Uint16("Offset", request.Offset), zap.Uint16("Count", request.Count))
	start := 1 + request.Offset
	end := 1 + request.Offset + request.Count
	results := h.InputRegisters[start:end]
	return &ReadInputRegistersResponse{Values: results}, nil
}

func (h *DefaultHandler) WriteSingleCoil(rctx context.Context, request *WriteSingleCoilRequest) (response *WriteSingleCoilResponse, err error) {
	h.logger.Debug("WriteSingleCoil", zap.Uint16("Offset", request.Offset), zap.Bool("Value", request.Value))
	h.Coils[request.Offset+1] = request.Value
	return &WriteSingleCoilResponse{
		Offset: request.Offset,
		Value:  request.Value,
	}, nil
}

func (h *DefaultHandler) WriteSingleRegister(ctx context.Context, request *WriteSingleRegisterRequest) (response *WriteSingleRegisterResponse, err error) {
	h.logger.Debug("WriteSingleRegister", zap.Uint16("Offset", request.Offset), zap.Uint16("Value", request.Value))
	h.HoldingRegisters[request.Offset+1] = request.Value
	return &WriteSingleRegisterResponse{
		Offset: request.Offset,
		Value:  request.Value,
	}, nil
}

func (h *DefaultHandler) WriteMultipleCoils(ctx context.Context, request *WriteMultipleCoilsRequest) (response *WriteMultipleCoilsResponse, err error) {
	h.logger.Debug("WriteMultipleCoils", zap.Uint16("Offset", request.Offset), zap.Bools("Values", request.Values))
	start := 1 + request.Offset
	for i, v := range request.Values {
		h.Coils[start+uint16(i)] = v
	}
	return &WriteMultipleCoilsResponse{
		Offset: request.Offset,
		Count:  uint16(len(request.Values)),
	}, nil
}

func (h *DefaultHandler) WriteMultipleRegisters(ctx context.Context, request *WriteMultipleRegistersRequest) (response *WriteMultipleRegistersResponse, err error) {
	h.logger.Debug("WriteMultipleRegisters", zap.Uint16("Offset", request.Offset), zap.Uint16s("Values", request.Values))
	start := 1 + request.Offset
	for i, v := range request.Values {
		h.HoldingRegisters[start+uint16(i)] = v
	}
	return &WriteMultipleRegistersResponse{
		Offset: request.Offset,
		Count:  uint16(len(request.Values)),
	}, nil
}
