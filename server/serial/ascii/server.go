package server

import (
	"errors"
	"io"

	sp "github.com/goburrow/serial"
	"github.com/rinzlerlabs/gomodbus/server"
	"github.com/rinzlerlabs/gomodbus/server/serial"
	"github.com/rinzlerlabs/gomodbus/transport/serial/ascii"
	"go.uber.org/zap"
)

func NewModbusServer(logger *zap.Logger, port sp.Port, serverAddress uint16) (server.ModbusServer, error) {
	handler := server.NewDefaultHandler(logger, 65535, 65535, 65535, 65535)
	return NewModbusServerWithHandler(logger, port, serverAddress, handler)
}

func NewModbusServerWithHandler(logger *zap.Logger, port sp.Port, serverAddress uint16, handler server.RequestHandler) (server.ModbusServer, error) {
	if handler == nil {
		return nil, errors.New("handler is required")
	}

	return serial.NewModbusSerialServerWithHandler(logger, serverAddress, handler, ascii.NewModbusTransport(port, logger))

}

// newModbusServerWithHandler creates a new Modbus ASCII server with a io.ReadWriter stream instead of an explicit port, for testing purposes, and a RequestHandler.
func newModbusServerWithHandler(logger *zap.Logger, stream io.ReadWriteCloser, serverAddress uint16, handler server.RequestHandler) (serial.ModbusSerialServer, error) {
	if handler == nil {
		return nil, errors.New("handler is required")
	}

	return serial.NewModbusSerialServerWithHandler(logger, serverAddress, handler, ascii.NewModbusTransport(stream, logger))
}
