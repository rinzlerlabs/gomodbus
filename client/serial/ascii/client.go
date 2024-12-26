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

func NewModbusClient(logger *zap.Logger, settings *sp.Config, responseTimeout time.Duration) (client.ModbusClient, error) {
	return NewModbusClientWithContext(context.Background(), logger, settings, responseTimeout)
}

func NewModbusClientWithContext(ctx context.Context, logger *zap.Logger, settings *sp.Config, responseTimeout time.Duration) (client.ModbusClient, error) {
	port, err := sp.Open(settings)
	if err != nil {
		return nil, err
	}
	transport := ascii.NewModbusTransport(port, logger)
	requestCreator := serial.NewSerialRequestCreator(ascii.NewModbusTransaction, ascii.NewModbusFrame)
	return client.NewModbusClient(ctx, logger, transport, requestCreator, responseTimeout), nil
}

func newModbusClient(logger *zap.Logger, stream io.ReadWriteCloser, responseTimeout time.Duration) client.ModbusClient {
	ctx := context.Background()
	transport := ascii.NewModbusTransport(stream, logger)
	requestCreator := serial.NewSerialRequestCreator(ascii.NewModbusTransaction, ascii.NewModbusFrame)
	return client.NewModbusClient(ctx, logger, transport, requestCreator, responseTimeout)
}
