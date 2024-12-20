package client

import (
	"context"
	"io"
	"time"

	"github.com/goburrow/serial"
	"github.com/rinzlerlabs/gomodbus/transport/serial/rtu"
	"go.uber.org/zap"
)

func NewModbusRTUClient(logger *zap.Logger, port serial.Port, responseTimeout time.Duration) ModbusClient {
	return NewModbusRTUClientWithContext(context.Background(), logger, port, responseTimeout)
}

func NewModbusRTUClientWithContext(ctx context.Context, logger *zap.Logger, port serial.Port, responseTimeout time.Duration) ModbusClient {
	return &modbusClient{
		transport:       rtu.NewModbusTransport(port, logger),
		logger:          logger,
		ctx:             ctx,
		responseTimeout: responseTimeout,
		requestCreator: &serialRequestCreator{
			newModbusFrame:    rtu.NewModbusFrame,
			createTransaction: rtu.NewModbusTransaction,
		},
	}
}

func newModbusRTUClient(logger *zap.Logger, stream io.ReadWriteCloser, responseTimeout time.Duration) ModbusClient {
	return &modbusClient{
		transport:       rtu.NewModbusTransport(stream, logger),
		logger:          logger,
		ctx:             context.Background(),
		responseTimeout: responseTimeout,
		requestCreator: &serialRequestCreator{
			newModbusFrame:    rtu.NewModbusFrame,
			createTransaction: rtu.NewModbusTransaction,
		},
	}
}
