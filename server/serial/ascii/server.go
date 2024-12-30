package ascii

import (
	"errors"

	sp "github.com/goburrow/serial"
	"github.com/rinzlerlabs/gomodbus/server"
	"github.com/rinzlerlabs/gomodbus/server/serial"
	"github.com/rinzlerlabs/gomodbus/transport"
	"github.com/rinzlerlabs/gomodbus/transport/serial/ascii"
	"go.uber.org/zap"
)

func NewModbusServer(logger *zap.Logger, settings *sp.Config, serverAddress uint16) (server.ModbusServer, error) {
	handler := server.NewDefaultHandler(logger, server.DefaultCoilCount, server.DefaultDiscreteInputCount, server.DefaultHoldingRegisterCount, server.DefaultInputRegisterCount)
	return NewModbusServerWithHandler(logger, settings, serverAddress, handler)
}

func NewModbusServerWithHandler(logger *zap.Logger, settings *sp.Config, serverAddress uint16, handler server.RequestHandler) (server.ModbusServer, error) {
	if handler == nil {
		return nil, errors.New("handler is required")
	}

	transportCreator := func() (transport.Transport, error) {
		port, err := sp.Open(settings)
		if err != nil {
			logger.Error("Failed to open serial port", zap.Error(err))
			return nil, err
		}
		return ascii.NewModbusServerTransport(port, logger), nil
	}

	return serial.NewModbusSerialServerWithCreator(logger, settings, serverAddress, handler, transportCreator)
}
