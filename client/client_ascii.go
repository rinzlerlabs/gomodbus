package client

import (
	"context"
	"io"
	"time"

	"github.com/goburrow/serial"
	"github.com/rinzlerlabs/gomodbus/transport/serial/ascii"
	"go.uber.org/zap"
)

func NewModbusASCIIClient(logger *zap.Logger, port serial.Port, responseTimeout time.Duration) ModbusClient {
	return NewModbusASCIIClientWithContext(context.Background(), logger, port, responseTimeout)
}

func NewModbusASCIIClientWithContext(ctx context.Context, logger *zap.Logger, port serial.Port, responseTimeout time.Duration) ModbusClient {
	return &modbusClient{
		transport:       ascii.NewModbusTransport(port, logger),
		logger:          logger,
		ctx:             ctx,
		responseTimeout: responseTimeout,
		requestCreator: &serialRequestCreator{
			newModbusFrame:    ascii.NewModbusFrame,
			createTransaction: ascii.NewModbusTransaction,
		},
	}
}

func newModbusASCIIClient(logger *zap.Logger, stream io.ReadWriteCloser, responseTimeout time.Duration) ModbusClient {
	return &modbusClient{
		transport:       ascii.NewModbusTransport(stream, logger),
		logger:          logger,
		ctx:             context.Background(),
		responseTimeout: responseTimeout,
		requestCreator: &serialRequestCreator{
			newModbusFrame:    ascii.NewModbusFrame,
			createTransaction: ascii.NewModbusTransaction,
		},
	}
}
