package client

import (
	"context"
	"io"
	"net"
	"time"

	"github.com/rinzlerlabs/gomodbus/transport/network/tcp"
	"github.com/rinzlerlabs/gomodbus/transport/serial/rtu"
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

func newModbusTCPClient(logger *zap.Logger, stream io.ReadWriteCloser, responseTimeout time.Duration) ModbusClient {
	return &modbusClient{
		transport:       rtu.NewModbusTransport(stream, logger),
		logger:          logger,
		ctx:             context.Background(),
		responseTimeout: responseTimeout,
		requestCreator: &networkRequestCreator{
			newModbusFrame:    tcp.NewModbusFrame,
			createTransaction: tcp.NewModbusTransaction,
		},
	}
}
