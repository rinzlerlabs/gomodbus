package network

import (
	"context"
	"net"
	"time"

	"github.com/rinzlerlabs/gomodbus/client"
	settings "github.com/rinzlerlabs/gomodbus/settings/network"
	transport "github.com/rinzlerlabs/gomodbus/transport/network"
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
	dialer := net.Dialer{
		Timeout:   settings.DialTimeout,
		KeepAlive: settings.KeepAlive,
	}
	dialContext, cancelFunc := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
	defer cancelFunc()
	conn, err := dialer.DialContext(dialContext, settings.Endpoint.Scheme, settings.Endpoint.Host)
	if err != nil {
		logger.Error("Failed to connect to endpoint", zap.String("endpoint", settings.Endpoint.String()), zap.Error(err))
		return nil, err
	}
	return client.NewModbusClient(ctx, logger, transport.NewModbusClientTransport(conn, logger, settings.ResponseTimeout)), nil
}
