package rtu

import (
	"context"
	"io"
	"time"

	sp "github.com/goburrow/serial"
	"github.com/rinzlerlabs/gomodbus/client"
	"github.com/rinzlerlabs/gomodbus/client/serial"
	"github.com/rinzlerlabs/gomodbus/transport/serial/rtu"
	"go.uber.org/zap"
)

func NewModbusRTUClient(logger *zap.Logger, port sp.Port, responseTimeout time.Duration) client.ModbusClient {
	return NewModbusRTUClientWithContext(context.Background(), logger, port, responseTimeout)
}

func NewModbusRTUClientWithContext(ctx context.Context, logger *zap.Logger, port sp.Port, responseTimeout time.Duration) client.ModbusClient {
	return client.NewModbusClient(ctx, logger, rtu.NewModbusClientTransport(port, logger), serial.NewSerialRequestCreator(rtu.NewModbusTransaction, rtu.NewModbusFrame), responseTimeout)
}

func newModbusRTUClient(logger *zap.Logger, stream io.ReadWriteCloser, responseTimeout time.Duration) client.ModbusClient {
	ctx := context.Background()
	return client.NewModbusClient(ctx, logger, rtu.NewModbusClientTransport(stream, logger), serial.NewSerialRequestCreator(rtu.NewModbusTransaction, rtu.NewModbusFrame), responseTimeout)
}
