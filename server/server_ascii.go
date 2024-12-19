package server

import (
	"context"
	"errors"
	"io"

	"github.com/goburrow/serial"
	"github.com/rinzlerlabs/gomodbus/transport/serial/ascii"
	"go.uber.org/zap"
)

func NewModbusASCIIServer(logger *zap.Logger, port serial.Port, serverAddress uint16) (ModbusServer, error) {
	handler := NewDefaultHandler(logger, 65535, 65535, 65535, 65535)
	return NewModbusASCIIServerWithHandler(logger, port, serverAddress, handler)
}

func NewModbusASCIIServerWithHandler(logger *zap.Logger, port serial.Port, serverAddress uint16, handler RequestHandler) (ModbusServer, error) {
	if handler == nil {
		return nil, errors.New("handler is required")
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &modbusSerialServer{
		logger:    logger,
		handler:   handler,
		cancelCtx: ctx,
		cancel:    cancel,
		address:   serverAddress,
		transport: ascii.NewModbusASCIITransport(port, logger),
	}, nil
}

// newModbusASCIIServerWithHandler creates a new Modbus ASCII server with a io.ReadWriter stream instead of an explicit port, for testing purposes, and a RequestHandler.
func newModbusASCIIServerWithHandler(logger *zap.Logger, stream io.ReadWriteCloser, serverAddress uint16, handler RequestHandler) (*modbusSerialServer, error) {
	if handler == nil {
		return nil, errors.New("handler is required")
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &modbusSerialServer{
		logger:    logger,
		handler:   handler,
		cancelCtx: ctx,
		cancel:    cancel,
		address:   serverAddress,
		transport: ascii.NewModbusASCIITransport(stream, logger),
	}, nil
}
