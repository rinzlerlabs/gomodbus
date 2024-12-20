package ascii

import (
	"context"
	"io"
	"time"

	sp "github.com/goburrow/serial"
	"github.com/rinzlerlabs/gomodbus/client"
	"github.com/rinzlerlabs/gomodbus/client/serial"
	"github.com/rinzlerlabs/gomodbus/transport/serial/ascii"
	"go.uber.org/zap"
)

func NewModbusASCIIClient(logger *zap.Logger, port sp.Port, responseTimeout time.Duration) client.ModbusClient {
	return NewModbusASCIIClientWithContext(context.Background(), logger, port, responseTimeout)
}

func NewModbusASCIIClientWithContext(ctx context.Context, logger *zap.Logger, port sp.Port, responseTimeout time.Duration) client.ModbusClient {
	return client.NewModbusClient(ctx, logger, ascii.NewModbusTransport(port, logger), serial.NewSerialRequestCreator(ascii.NewModbusTransaction, ascii.NewModbusFrame), responseTimeout)
}

func newModbusASCIIClient(logger *zap.Logger, stream io.ReadWriteCloser, responseTimeout time.Duration) client.ModbusClient {
	ctx := context.Background()
	return client.NewModbusClient(ctx, logger, ascii.NewModbusTransport(stream, logger), serial.NewSerialRequestCreator(ascii.NewModbusTransaction, ascii.NewModbusFrame), responseTimeout)
}
