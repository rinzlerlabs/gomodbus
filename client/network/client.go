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

func NewModbusClient(logger *zap.Logger, endpoint string, responseTimeout time.Duration) (client.ModbusClient, error) {
	return NewModbusClientWithContext(context.Background(), logger, endpoint, responseTimeout)
}

func NewModbusClientWithContext(ctx context.Context, logger *zap.Logger, endpoint string, responseTimeout time.Duration) (client.ModbusClient, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		logger.Error("Failed to parse endpoint", zap.String("endpoint", endpoint), zap.Error(err))
		return nil, err
	}
	conn, err := net.Dial(u.Scheme, u.Host)
	if err != nil {
		logger.Error("Failed to connect to endpoint", zap.String("endpoint", endpoint), zap.Error(err))
		return nil, err
	}
	return client.NewModbusClient(ctx, logger, transport.NewModbusTransport(conn, logger), NewNetworkRequestCreator(transport.NewModbusTransaction, transport.NewModbusFrame), responseTimeout), nil
}

func newModbusClient(logger *zap.Logger, stream transport.ReadWriteCloseRemoteAddresser, responseTimeout time.Duration) client.ModbusClient {
	ctx := context.Background()
	return client.NewModbusClient(ctx, logger, transport.NewModbusTransport(stream, logger), NewNetworkRequestCreator(transport.NewModbusTransaction, transport.NewModbusFrame), responseTimeout)
}
