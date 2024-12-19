package server

import (
	"context"
	"errors"

	"github.com/rinzlerlabs/gomodbus/transport/network/tcp"
	"go.uber.org/zap"
)

func NewModbusTCPServer(logger *zap.Logger, endpoint string) (ModbusServer, error) {
	handler := NewDefaultHandler(logger, 65535, 65535, 65535, 65535)
	return NewModbusTCPServerWithHandler(logger, endpoint, handler)
}

func NewModbusTCPServerWithHandler(logger *zap.Logger, endpoint string, handler RequestHandler) (ModbusServer, error) {
	if handler == nil {
		return nil, errors.New("handler is required")
	}
	transport, err := tcp.NewModbusTCPTransport(endpoint, logger)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &modbusServer{
		logger:    logger,
		handler:   handler,
		cancelCtx: ctx,
		cancel:    cancel,
		transport: transport,
	}, nil
}
