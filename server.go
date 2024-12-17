package gomodbus

import (
	"bufio"
	"context"
	"io"
	"sync"
	"time"

	"go.uber.org/zap"
)

type ModbusServer interface {
	Start() error
	Stop() error
}

type modbusServer struct {
	handler             RequestHandler
	mu                  sync.Mutex
	cancelCtx           context.Context
	cancel              context.CancelFunc
	logger              *zap.Logger
	responseCreatorFunc func(uint16, byte, ModbusResponse) ApplicationDataUnit
	responseFormatter   func(ApplicationDataUnit) []byte
	responseWriter      io.Writer
}

type modbusSerialServer struct {
	modbusServer
	address uint16
	port    io.ReadWriteCloser
	reader  *bufio.Reader
}

type ModbusTCPServer struct {
	Endpoints []string
	Timeout   time.Duration
}

type ModbusUDPServer struct {
	Endpoints []string
	Timeout   time.Duration
}

func (s *modbusServer) writeResponse(adu ApplicationDataUnit) error {
	data := s.responseFormatter(adu)
	n, err := s.responseWriter.Write(data)
	if err != nil {
		return err
	}
	if n != len(data) {
		return io.ErrShortWrite
	}
	return nil
}

func (s *modbusSerialServer) handlePacket(packet ApplicationDataUnit) error {
	if packet.Address() != s.address {
		s.logger.Debug("Received packet with incorrect address, discarding packet", zap.Any("packet", packet))
		return ErrNotOurAddress
	}
	switch packet.PDU().Function {
	case FunctionCodeReadCoils:
		// Read Coils
		req, err := NewReadCoilsRequest(packet)
		if err != nil {
			s.logger.Error("Failed to parse Read Coils request, discarding packet", zap.Error(err))
			return err
		}
		s.logger.Debug("Received Read Coils request", zap.Any("request", req))
		result, err := s.handler.ReadCoils(s.cancelCtx, req)
		if err != nil {
			s.logger.Error("Failed to handle Read Coils request", zap.Error(err))
			return err
		}
		response := s.responseCreatorFunc(packet.Address(), packet.PDU().Function, result)
		s.logger.Debug("Sending Read Coils response", zap.Any("response", response))

		return s.writeResponse(response)
	case FunctionCodeReadDiscreteInputs:
		// Read Discrete Inputs
		req, err := NewReadDiscreteInputsRequest(packet)
		if err != nil {
			s.logger.Error("Failed to parse Read Discrete Inputs request, discarding packet", zap.Error(err))
			return err
		}
		s.logger.Debug("Received Read Discrete Inputs request", zap.Any("request", req))
		result, err := s.handler.ReadDiscreteInputs(s.cancelCtx, req)
		if err != nil {
			s.logger.Error("Failed to handle Read Discrete Inputs request", zap.Error(err))
			return err
		}
		response := s.responseCreatorFunc(packet.Address(), packet.PDU().Function, result)
		s.logger.Debug("Sending Read Discrete Inputs response", zap.Any("response", response))

		return s.writeResponse(response)
	case FunctionCodeReadHoldingRegisters:
		// Read Holding Registers
		req, err := NewReadHoldingRegistersRequest(packet)
		if err != nil {
			s.logger.Error("Failed to parse Read Holding Registers request, discarding packet", zap.Error(err))
			return err
		}
		s.logger.Debug("Received Read Holding Registers request", zap.Any("request", req))
		result, err := s.handler.ReadHoldingRegisters(s.cancelCtx, req)
		if err != nil {
			s.logger.Error("Failed to handle Read Holding Registers request", zap.Error(err))
			return err
		}
		response := s.responseCreatorFunc(packet.Address(), packet.PDU().Function, result)
		s.logger.Debug("Sending Read Holding Registers response", zap.Any("response", response))

		return s.writeResponse(response)
	case FunctionCodeReadInputRegisters:
		// Read Input Registers
		req, err := NewReadInputRegistersRequest(packet)
		if err != nil {
			s.logger.Error("Failed to parse Read Input Registers request, discarding packet", zap.Error(err))
			return err
		}
		s.logger.Debug("Received Read Input Registers request", zap.Any("request", req))
		result, err := s.handler.ReadInputRegisters(s.cancelCtx, req)
		if err != nil {
			s.logger.Error("Failed to handle Read Input Registers request", zap.Error(err))
			return err
		}
		response := s.responseCreatorFunc(packet.Address(), packet.PDU().Function, result)
		s.logger.Debug("Sending Read Input Registers response", zap.Any("response", response))

		return s.writeResponse(response)
	case FunctionCodeWriteSingleCoil:
		// Write Single Coil
		req, err := NewWriteSingleCoilRequest(packet)
		if err != nil {
			s.logger.Error("Failed to parse Write Single Coil request, discarding packet", zap.Error(err))
			return err
		}
		s.logger.Debug("Received Write Single Coil request", zap.Any("request", req))
		result, err := s.handler.WriteSingleCoil(s.cancelCtx, req)
		if err != nil {
			s.logger.Error("Failed to handle Write Single Coil request", zap.Error(err))
			return err
		}
		response := s.responseCreatorFunc(packet.Address(), packet.PDU().Function, result)
		s.logger.Debug("Sending Write Single Coil response", zap.Any("response", response))

		return s.writeResponse(response)
	case FunctionCodeWriteSingleRegister:
		// Write Single Register
		req, err := NewWriteSingleRegisterRequest(packet)
		if err != nil {
			s.logger.Error("Failed to parse Write Single Register request, discarding packet", zap.Error(err))
			return err
		}
		s.logger.Debug("Received Write Single Register request", zap.Any("request", req))
		result, err := s.handler.WriteSingleRegister(s.cancelCtx, req)
		if err != nil {
			s.logger.Error("Failed to handle Write Single Register request", zap.Error(err))
			return err
		}
		response := s.responseCreatorFunc(packet.Address(), packet.PDU().Function, result)
		s.logger.Debug("Sending Write Single Register response", zap.Any("response", response))

		return s.writeResponse(response)
	case FunctionCodeWriteMultipleCoils:
		// Write Multiple Coils
		req, err := NewWriteMultipleCoilsRequest(packet)
		if err != nil {
			s.logger.Error("Failed to parse Write Multiple Coils request, discarding packet", zap.Error(err))
			return err
		}
		s.logger.Debug("Received Write Multiple Coils request", zap.Any("request", req))
		result, err := s.handler.WriteMultipleCoils(s.cancelCtx, req)
		if err != nil {
			s.logger.Error("Failed to handle Write Multiple Coils request", zap.Error(err))
			return err
		}
		response := s.responseCreatorFunc(packet.Address(), packet.PDU().Function, result)
		s.logger.Debug("Sending Write Multiple Coils response", zap.Any("response", response))

		return s.writeResponse(response)
	case FunctionCodeWriteMultipleRegisters:
		// Write Multiple Registers
		req, err := NewWriteMultipleRegistersRequest(packet)
		if err != nil {
			s.logger.Error("Failed to parse Write Multiple Registers request, discarding packet", zap.Error(err))
			return err
		}
		s.logger.Debug("Received Write Multiple Registers request", zap.Any("request", req))
		result, err := s.handler.WriteMultipleRegisters(s.cancelCtx, req)
		if err != nil {
			s.logger.Error("Failed to handle Write Multiple Registers request", zap.Error(err))
			return err
		}
		response := s.responseCreatorFunc(packet.Address(), packet.PDU().Function, result)
		s.logger.Debug("Sending Write Multiple Registers response", zap.Any("response", response))

		return s.writeResponse(response)
	default:
		s.logger.Debug("Received packet with unknown function code, discarding packet", zap.Any("packet", packet))
		return ErrUnknownFunctionCode
	}
}
