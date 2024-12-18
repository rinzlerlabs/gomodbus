package server

import (
	"github.com/rinzlerlabs/gomodbus/common"
	"github.com/rinzlerlabs/gomodbus/data"
	"go.uber.org/zap"
)

// RequestHandler is the interface that wraps the basic Modbus functions.
// TODO: Merge single and multiple requests into one.
type RequestHandler interface {
	Handle(op data.ModbusFrame) error
	ReadCoils(request *data.ReadCoilsRequest) (response *data.ReadCoilsResponse, err error)
	ReadDiscreteInputs(request *data.ReadDiscreteInputsRequest) (response *data.ReadDiscreteInputsResponse, err error)
	ReadHoldingRegisters(request *data.ReadHoldingRegistersRequest) (response *data.ReadHoldingRegistersResponse, err error)
	ReadInputRegisters(request *data.ReadInputRegistersRequest) (response *data.ReadInputRegistersResponse, err error)
	WriteSingleCoil(equest *data.WriteSingleCoilRequest) (response *data.WriteSingleCoilResponse, err error)
	WriteSingleRegister(request *data.WriteSingleRegisterRequest) (response *data.WriteSingleRegisterResponse, err error)
	WriteMultipleCoils(request *data.WriteMultipleCoilsRequest) (response *data.WriteMultipleCoilsResponse, err error)
	WriteMultipleRegisters(request *data.WriteMultipleRegistersRequest) (response *data.WriteMultipleRegistersResponse, err error)
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

func (h *DefaultHandler) Handle(request data.ModbusFrame) error {
	h.logger.Debug("Received packet", zap.Any("packet", request))
	var response *data.ProtocolDataUnit
	switch request.PDU().Function {
	case data.ReadCoils:
		// Read Coils
		req, err := newReadCoilsRequest(request)
		if err != nil {
			h.logger.Error("Failed to parse Read Coils request, discarding packet", zap.Error(err))
			return err
		}
		h.logger.Info("Received Read Coils request", zap.Any("request", req))
		result, err := h.ReadCoils(req)
		if err != nil {
			h.logger.Error("Failed to handle Read Coils request", zap.Error(err))
			return err
		}
		h.logger.Debug("Read coil successful", zap.Any("result", result))
		response = &data.ProtocolDataUnit{Function: data.ReadCoils, Data: result.Bytes()}
	case data.ReadDiscreteInputs:
		// Read Discrete Inputs
		req, err := newReadDiscreteInputsRequest(request)
		if err != nil {
			h.logger.Error("Failed to parse Read Discrete Inputs request, discarding packet", zap.Error(err))
			return err
		}
		h.logger.Info("Received Read Discrete Inputs request", zap.Any("request", req))
		result, err := h.ReadDiscreteInputs(req)
		if err != nil {
			h.logger.Error("Failed to handle Read Discrete Inputs request", zap.Error(err))
			return err
		}
		h.logger.Debug("Read Discrete Inputs successful", zap.Any("result", result))
		response = &data.ProtocolDataUnit{Function: data.ReadDiscreteInputs, Data: result.Bytes()}
	case data.ReadHoldingRegisters:
		// Read Holding Registers
		req, err := newReadHoldingRegistersRequest(request)
		if err != nil {
			h.logger.Error("Failed to parse Read Holding Registers request, discarding packet", zap.Error(err))
			return err
		}
		h.logger.Info("Received Read Holding Registers request", zap.Any("request", req))
		result, err := h.ReadHoldingRegisters(req)
		if err != nil {
			h.logger.Error("Failed to handle Read Holding Registers request", zap.Error(err))
			return err
		}
		h.logger.Debug("Read Holding Registers successful", zap.Any("result", result))
		response = &data.ProtocolDataUnit{Function: data.ReadHoldingRegisters, Data: result.Bytes()}
	case data.ReadInputRegisters:
		// Read Input Registers
		req, err := newReadInputRegistersRequest(request)
		if err != nil {
			h.logger.Error("Failed to parse Read Input Registers request, discarding packet", zap.Error(err))
			return err
		}
		h.logger.Info("Received Read Input Registers request", zap.Any("request", req))
		result, err := h.ReadInputRegisters(req)
		if err != nil {
			h.logger.Error("Failed to handle Read Input Registers request", zap.Error(err))
			return err
		}
		h.logger.Debug("Read Input Registers successful", zap.Any("result", result))
		response = &data.ProtocolDataUnit{Function: data.ReadInputRegisters, Data: result.Bytes()}
	case data.WriteSingleCoil:
		// Write Single Coil
		req, err := newWriteSingleCoilRequest(request)
		if err != nil {
			h.logger.Error("Failed to parse Write Single Coil request, discarding packet", zap.Error(err))
			return err
		}
		h.logger.Info("Received Write Single Coil request", zap.Any("request", req))
		result, err := h.WriteSingleCoil(req)
		if err != nil {
			h.logger.Error("Failed to handle Write Single Coil request", zap.Error(err))
			return err
		}
		h.logger.Debug("Write Single Coil successful", zap.Any("result", result))
		response = &data.ProtocolDataUnit{Function: data.WriteSingleCoil, Data: result.Bytes()}
	case data.WriteSingleRegister:
		// Write Single Register
		req, err := newWriteSingleRegisterRequest(request)
		if err != nil {
			h.logger.Error("Failed to parse Write Single Register request, discarding packet", zap.Error(err))
			return err
		}
		h.logger.Info("Received Write Single Register request", zap.Any("request", req))
		result, err := h.WriteSingleRegister(req)
		if err != nil {
			h.logger.Error("Failed to handle Write Single Register request", zap.Error(err))
			return err
		}
		h.logger.Debug("Write Single Register successful", zap.Any("result", result))
		response = &data.ProtocolDataUnit{Function: data.WriteSingleRegister, Data: result.Bytes()}
	case data.WriteMultipleCoils:
		// Write Multiple Coils
		req, err := newWriteMultipleCoilsRequest(request)
		if err != nil {
			h.logger.Error("Failed to parse Write Multiple Coils request, discarding packet", zap.Error(err))
			return err
		}
		h.logger.Info("Received Write Multiple Coils request", zap.Any("request", req))
		result, err := h.WriteMultipleCoils(req)
		if err != nil {
			h.logger.Error("Failed to handle Write Multiple Coils request", zap.Error(err))
			return err
		}
		h.logger.Debug("Write Multiple Coils successful", zap.Any("result", result))
		response = &data.ProtocolDataUnit{Function: data.WriteMultipleCoils, Data: result.Bytes()}
	case data.WriteMultipleRegisters:
		// Write Multiple Registers
		req, err := newWriteMultipleRegistersRequest(request)
		if err != nil {
			h.logger.Error("Failed to parse Write Multiple Registers request, discarding packet", zap.Error(err))
			return err
		}
		h.logger.Info("Received Write Multiple Registers request", zap.Any("request", req))
		result, err := h.WriteMultipleRegisters(req)
		if err != nil {
			h.logger.Error("Failed to handle Write Multiple Registers request", zap.Error(err))
			return err
		}
		h.logger.Debug("Write Multiple Registers successful", zap.Any("result", result))
		response = &data.ProtocolDataUnit{Function: data.WriteMultipleRegisters, Data: result.Bytes()}
	default:
		h.logger.Debug("Received packet with unknown function code, discarding packet", zap.Any("packet", request))
		return common.ErrUnknownFunctionCode
	}
	h.logger.Info("Request handled successfully", zap.Any("response", response))
	// Send response
	return request.SendResponse(response)
}

func (h *DefaultHandler) ReadCoils(request *data.ReadCoilsRequest) (response *data.ReadCoilsResponse, err error) {
	h.logger.Debug("ReadCoils", zap.Uint16("Offset", request.Offset), zap.Uint16("Count", request.Count))
	start := 1 + request.Offset
	end := 1 + request.Offset + request.Count
	results := h.Coils[start:end]
	return &data.ReadCoilsResponse{Values: results}, nil
}

func (h *DefaultHandler) ReadDiscreteInputs(request *data.ReadDiscreteInputsRequest) (response *data.ReadDiscreteInputsResponse, err error) {
	h.logger.Debug("ReadDiscreteInputs", zap.Uint16("Offset", request.Offset), zap.Uint16("Count", request.Count))
	start := 1 + request.Offset
	end := 1 + request.Offset + request.Count
	results := h.DiscreteInputs[start:end]
	return &data.ReadDiscreteInputsResponse{Values: results}, nil
}

func (h *DefaultHandler) ReadHoldingRegisters(request *data.ReadHoldingRegistersRequest) (response *data.ReadHoldingRegistersResponse, err error) {
	h.logger.Debug("ReadHoldingRegisters", zap.Uint16("Offset", request.Offset), zap.Uint16("Count", request.Count))
	start := 1 + request.Offset
	end := 1 + request.Offset + request.Count
	results := h.HoldingRegisters[start:end]
	return &data.ReadHoldingRegistersResponse{Values: results}, nil
}

func (h *DefaultHandler) ReadInputRegisters(request *data.ReadInputRegistersRequest) (response *data.ReadInputRegistersResponse, err error) {
	h.logger.Debug("ReadInputRegisters", zap.Uint16("Offset", request.Offset), zap.Uint16("Count", request.Count))
	start := 1 + request.Offset
	end := 1 + request.Offset + request.Count
	results := h.InputRegisters[start:end]
	return &data.ReadInputRegistersResponse{Values: results}, nil
}

func (h *DefaultHandler) WriteSingleCoil(request *data.WriteSingleCoilRequest) (response *data.WriteSingleCoilResponse, err error) {
	h.logger.Debug("WriteSingleCoil", zap.Uint16("Offset", request.Offset), zap.Bool("Value", request.Value))
	h.Coils[request.Offset+1] = request.Value
	return &data.WriteSingleCoilResponse{
		Offset: request.Offset,
		Value:  request.Value,
	}, nil
}

func (h *DefaultHandler) WriteSingleRegister(request *data.WriteSingleRegisterRequest) (response *data.WriteSingleRegisterResponse, err error) {
	h.logger.Debug("WriteSingleRegister", zap.Uint16("Offset", request.Offset), zap.Uint16("Value", request.Value))
	h.HoldingRegisters[request.Offset+1] = request.Value
	return &data.WriteSingleRegisterResponse{
		Offset: request.Offset,
		Value:  request.Value,
	}, nil
}

func (h *DefaultHandler) WriteMultipleCoils(request *data.WriteMultipleCoilsRequest) (response *data.WriteMultipleCoilsResponse, err error) {
	h.logger.Debug("WriteMultipleCoils", zap.Uint16("Offset", request.Offset), zap.Bools("Values", request.Values))
	start := 1 + request.Offset
	for i, v := range request.Values {
		h.Coils[start+uint16(i)] = v
	}
	return &data.WriteMultipleCoilsResponse{
		Offset: request.Offset,
		Count:  uint16(len(request.Values)),
	}, nil
}

func (h *DefaultHandler) WriteMultipleRegisters(request *data.WriteMultipleRegistersRequest) (response *data.WriteMultipleRegistersResponse, err error) {
	h.logger.Debug("WriteMultipleRegisters", zap.Uint16("Offset", request.Offset), zap.Uint16s("Values", request.Values))
	start := 1 + request.Offset
	for i, v := range request.Values {
		h.HoldingRegisters[start+uint16(i)] = v
	}
	return &data.WriteMultipleRegistersResponse{
		Offset: request.Offset,
		Count:  uint16(len(request.Values)),
	}, nil
}
