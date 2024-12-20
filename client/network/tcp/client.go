package tcp

import (
	"context"
	"net"
	"time"

	"github.com/rinzlerlabs/gomodbus/client"
	"github.com/rinzlerlabs/gomodbus/client/network"
	"github.com/rinzlerlabs/gomodbus/transport/network/tcp"
	"go.uber.org/zap"
)

func NewModbusTCPClient(logger *zap.Logger, endpoint string, responseTimeout time.Duration) (client.ModbusClient, error) {
	return NewModbusTCPClientWithContext(context.Background(), logger, endpoint, responseTimeout)
}

func NewModbusTCPClientWithContext(ctx context.Context, logger *zap.Logger, endpoint string, responseTimeout time.Duration) (client.ModbusClient, error) {
	conn, err := net.Dial("tcp", endpoint)
	if err != nil {
		logger.Error("Failed to connect to endpoint", zap.String("endpoint", endpoint), zap.Error(err))
		panic(err)
	}
	return client.NewModbusClient(ctx, logger, tcp.NewModbusTransport(conn, logger), network.NewNetworkRequestCreator(tcp.NewModbusTransaction, tcp.NewModbusFrame), responseTimeout), nil
}

func newModbusTCPClient(logger *zap.Logger, stream tcp.ReadWriteCloseRemoteAddresser, responseTimeout time.Duration) client.ModbusClient {
	ctx := context.Background()
	return client.NewModbusClient(ctx, logger, tcp.NewModbusTransport(stream, logger), network.NewNetworkRequestCreator(tcp.NewModbusTransaction, tcp.NewModbusFrame), responseTimeout)
}
