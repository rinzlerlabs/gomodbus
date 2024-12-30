package network

import (
	"context"
	"net"
	"net/url"
	"time"

	"github.com/rinzlerlabs/gomodbus/client"
	transport "github.com/rinzlerlabs/gomodbus/transport/network"
	"go.uber.org/zap"
)

func NewModbusClient(logger *zap.Logger, uri string) (client.ModbusClient, error) {
	return NewModbusClientWithContext(context.Background(), logger, uri)
}

func NewModbusClientWithContext(ctx context.Context, logger *zap.Logger, uri string) (client.ModbusClient, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	settings, err := NewClientSettingsFromURI(u)
	return NewModbusClientFromSettingsWithContext(ctx, logger, settings)
}

func NewModbusClientFromSettings(logger *zap.Logger, settings *ClientSettings) (client.ModbusClient, error) {
	return NewModbusClientFromSettingsWithContext(context.Background(), logger, settings)
}

func NewModbusClientFromSettingsWithContext(ctx context.Context, logger *zap.Logger, settings *ClientSettings) (client.ModbusClient, error) {
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
	return client.NewModbusClient(ctx, logger, transport.NewModbusTransport(conn, logger), NewNetworkRequestCreator(transport.NewModbusTransaction, transport.NewModbusFrame), settings.ResponseTimeout), nil
}

func newModbusClient(logger *zap.Logger, stream transport.ReadWriteCloseRemoteAddresser, responseTimeout time.Duration) client.ModbusClient {
	ctx := context.Background()
	return client.NewModbusClient(ctx, logger, transport.NewModbusTransport(stream, logger), NewNetworkRequestCreator(transport.NewModbusTransaction, transport.NewModbusFrame), responseTimeout)
}
