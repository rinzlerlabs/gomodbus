package ascii

import (
	sp "github.com/goburrow/serial"
	"github.com/rinzlerlabs/gomodbus/common"
	"github.com/rinzlerlabs/gomodbus/server"
	"github.com/rinzlerlabs/gomodbus/server/serial"
	settings "github.com/rinzlerlabs/gomodbus/settings/serial"
	"github.com/rinzlerlabs/gomodbus/transport/serial/ascii"
	"go.uber.org/zap"
)

func NewModbusServer(logger *zap.Logger, uri string) (server.ModbusServer, error) {
	settings, err := settings.NewServerSettingsFromURI(uri)
	if err != nil {
		return nil, err
	}
	return NewModbusServerFromSettings(logger, settings)
}

func NewModbusServerFromSettings(logger *zap.Logger, serverSettings *settings.ServerSettings) (server.ModbusServer, error) {
	handler := server.NewDefaultHandler(logger, server.DefaultCoilCount, server.DefaultDiscreteInputCount, server.DefaultHoldingRegisterCount, server.DefaultInputRegisterCount)
	return NewModbusServerWithHandler(logger, serverSettings, handler)
}

func NewModbusServerWithHandler(logger *zap.Logger, serverSettings *settings.ServerSettings, handler server.RequestHandler) (server.ModbusServer, error) {
	if handler == nil {
		return nil, common.ErrHandlerRequired
	}

	port, err := sp.Open(serverSettings.GetSerialPortConfig())
	if err != nil {
		logger.Error("Failed to open serial port", zap.Error(err))
		return nil, err
	}
	transport := ascii.NewModbusServerTransport(port, logger)

	return serial.NewModbusSerialServerWithTransport(logger, serverSettings, handler, transport)
}
