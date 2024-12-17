package gomodbus

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

type ModbusServer interface {
	Start() error
	Stop() error
}

type modbusServer struct {
	handler   RequestHandler
	mu        sync.Mutex
	cancelCtx context.Context
	cancel    context.CancelFunc
	logger    *zap.Logger
}

func (s *modbusServer) handlePacket(op ModbusOperation) error {
	var result ModbusResponse
	switch op.Request().PDU().Function {
	case FunctionCodeReadCoils:
		// Read Coils
		req, err := NewReadCoilsRequest(op.Request())
		if err != nil {
			s.logger.Error("Failed to parse Read Coils request, discarding packet", zap.Error(err))
			return err
		}
		s.logger.Debug("Received Read Coils request", zap.Any("request", req))
		result, err = s.handler.ReadCoils(s.cancelCtx, req)
		if err != nil {
			s.logger.Error("Failed to handle Read Coils request", zap.Error(err))
			return err
		}
		s.logger.Debug("Read coil successful", zap.Any("result", result))
	case FunctionCodeReadDiscreteInputs:
		// Read Discrete Inputs
		req, err := NewReadDiscreteInputsRequest(op.Request())
		if err != nil {
			s.logger.Error("Failed to parse Read Discrete Inputs request, discarding packet", zap.Error(err))
			return err
		}
		s.logger.Debug("Received Read Discrete Inputs request", zap.Any("request", req))
		result, err = s.handler.ReadDiscreteInputs(s.cancelCtx, req)
		if err != nil {
			s.logger.Error("Failed to handle Read Discrete Inputs request", zap.Error(err))
			return err
		}
		s.logger.Debug("Read Discrete Inputs successful", zap.Any("result", result))
	case FunctionCodeReadHoldingRegisters:
		// Read Holding Registers
		req, err := NewReadHoldingRegistersRequest(op.Request())
		if err != nil {
			s.logger.Error("Failed to parse Read Holding Registers request, discarding packet", zap.Error(err))
			return err
		}
		s.logger.Debug("Received Read Holding Registers request", zap.Any("request", req))
		result, err = s.handler.ReadHoldingRegisters(s.cancelCtx, req)
		if err != nil {
			s.logger.Error("Failed to handle Read Holding Registers request", zap.Error(err))
			return err
		}
		s.logger.Debug("Read Holding Registers successful", zap.Any("result", result))
	case FunctionCodeReadInputRegisters:
		// Read Input Registers
		req, err := NewReadInputRegistersRequest(op.Request())
		if err != nil {
			s.logger.Error("Failed to parse Read Input Registers request, discarding packet", zap.Error(err))
			return err
		}
		s.logger.Debug("Received Read Input Registers request", zap.Any("request", req))
		result, err = s.handler.ReadInputRegisters(s.cancelCtx, req)
		if err != nil {
			s.logger.Error("Failed to handle Read Input Registers request", zap.Error(err))
			return err
		}
		s.logger.Debug("Read Input Registers successful", zap.Any("result", result))
	case FunctionCodeWriteSingleCoil:
		// Write Single Coil
		req, err := NewWriteSingleCoilRequest(op.Request())
		if err != nil {
			s.logger.Error("Failed to parse Write Single Coil request, discarding packet", zap.Error(err))
			return err
		}
		s.logger.Debug("Received Write Single Coil request", zap.Any("request", req))
		result, err = s.handler.WriteSingleCoil(s.cancelCtx, req)
		if err != nil {
			s.logger.Error("Failed to handle Write Single Coil request", zap.Error(err))
			return err
		}
		s.logger.Debug("Write Single Coil successful", zap.Any("result", result))
	case FunctionCodeWriteSingleRegister:
		// Write Single Register
		req, err := NewWriteSingleRegisterRequest(op.Request())
		if err != nil {
			s.logger.Error("Failed to parse Write Single Register request, discarding packet", zap.Error(err))
			return err
		}
		s.logger.Debug("Received Write Single Register request", zap.Any("request", req))
		result, err = s.handler.WriteSingleRegister(s.cancelCtx, req)
		if err != nil {
			s.logger.Error("Failed to handle Write Single Register request", zap.Error(err))
			return err
		}
		s.logger.Debug("Write Single Register successful", zap.Any("result", result))
	case FunctionCodeWriteMultipleCoils:
		// Write Multiple Coils
		req, err := NewWriteMultipleCoilsRequest(op.Request())
		if err != nil {
			s.logger.Error("Failed to parse Write Multiple Coils request, discarding packet", zap.Error(err))
			return err
		}
		s.logger.Debug("Received Write Multiple Coils request", zap.Any("request", req))
		result, err = s.handler.WriteMultipleCoils(s.cancelCtx, req)
		if err != nil {
			s.logger.Error("Failed to handle Write Multiple Coils request", zap.Error(err))
			return err
		}
		s.logger.Debug("Write Multiple Coils successful", zap.Any("result", result))
	case FunctionCodeWriteMultipleRegisters:
		// Write Multiple Registers
		req, err := NewWriteMultipleRegistersRequest(op.Request())
		if err != nil {
			s.logger.Error("Failed to parse Write Multiple Registers request, discarding packet", zap.Error(err))
			return err
		}
		s.logger.Debug("Received Write Multiple Registers request", zap.Any("request", req))
		result, err = s.handler.WriteMultipleRegisters(s.cancelCtx, req)
		if err != nil {
			s.logger.Error("Failed to handle Write Multiple Registers request", zap.Error(err))
			return err
		}
		s.logger.Debug("Write Multiple Registers successful", zap.Any("result", result))
	default:
		s.logger.Debug("Received packet with unknown function code, discarding packet", zap.Any("packet", op))
		return ErrUnknownFunctionCode
	}
	return op.SendResponse(result)
}
