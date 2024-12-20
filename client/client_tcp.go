package client

import (
	"context"
	"net"
	"time"

	"github.com/rinzlerlabs/gomodbus/transport/network/tcp"
	"go.uber.org/zap"
)

func NewModbusTCPClient(logger *zap.Logger, endpoint string, responseTimeout time.Duration) (ModbusClient, error) {
	return NewModbusTCPClientWithContext(context.Background(), logger, endpoint, responseTimeout)
}

func NewModbusTCPClientWithContext(ctx context.Context, logger *zap.Logger, endpoint string, responseTimeout time.Duration) (ModbusClient, error) {
	conn, err := net.Dial("tcp", endpoint)
	if err != nil {
		logger.Error("Failed to connect to endpoint", zap.String("endpoint", endpoint), zap.Error(err))
		panic(err)
	}
	return &modbusClient{
		transport:       tcp.NewModbusTransport(conn, logger),
		logger:          logger,
		ctx:             ctx,
		responseTimeout: responseTimeout,
		requestCreator: &networkRequestCreator{
			newModbusFrame:    tcp.NewModbusFrame,
			createTransaction: tcp.NewModbusTransaction,
		},
	}, nil
}

func newModbusTCPClient(logger *zap.Logger, stream tcp.ReadWriteCloseRemoteAddresser, responseTimeout time.Duration) ModbusClient {
	return &modbusClient{
		transport:       tcp.NewModbusTransport(stream, logger),
		logger:          logger,
		ctx:             context.Background(),
		responseTimeout: responseTimeout,
		requestCreator: &networkRequestCreator{
			newModbusFrame:    tcp.NewModbusFrame,
			createTransaction: tcp.NewModbusTransaction,
		},
	}
}
