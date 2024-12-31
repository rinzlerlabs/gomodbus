package ascii

import (
	"context"

	sp "github.com/goburrow/serial"
	"github.com/rinzlerlabs/gomodbus/client"
	settings "github.com/rinzlerlabs/gomodbus/settings/serial"
	"github.com/rinzlerlabs/gomodbus/transport/serial/ascii"
	"go.uber.org/zap"
)

func NewModbusClient(logger *zap.Logger, uri string) (client.ModbusClient, error) {
	return NewModbusClientWithContext(context.Background(), logger, uri)
}

func NewModbusClientWithContext(ctx context.Context, logger *zap.Logger, uri string) (client.ModbusClient, error) {
	settings, err := settings.NewClientSettingsFromURI(uri)
	if err != nil {
		return nil, err
	}
	return NewModbusClientFromSettingsWithContext(ctx, logger, settings)
}

func NewModbusClientFromSettings(logger *zap.Logger, settings *settings.ClientSettings) (client.ModbusClient, error) {
	return NewModbusClientFromSettingsWithContext(context.Background(), logger, settings)
}

func NewModbusClientFromSettingsWithContext(ctx context.Context, logger *zap.Logger, settings *settings.ClientSettings) (client.ModbusClient, error) {
	config := settings.GetSerialPortConfig()
	port, err := sp.Open(config)
	if err != nil {
		return nil, err
	}
	t := ascii.NewModbusClientTransport(port, logger, settings.ResponseTimeout)
	return client.NewModbusClient(ctx, logger, t), nil
}
