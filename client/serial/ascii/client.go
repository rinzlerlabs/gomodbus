package ascii

import (
	"context"
	"io"
	"net/url"
	"time"

	sp "github.com/goburrow/serial"
	"github.com/rinzlerlabs/gomodbus/client"
	"github.com/rinzlerlabs/gomodbus/client/serial"
	"github.com/rinzlerlabs/gomodbus/transport/serial/ascii"
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
	settings, err := serial.NewClientSettingsFromURI(u)
	return NewModbusClientFromSettingsWithContext(ctx, logger, settings)
}

func NewModbusClientFromSettings(logger *zap.Logger, settings *serial.ClientSettings) (client.ModbusClient, error) {
	return NewModbusClientFromSettingsWithContext(context.Background(), logger, settings)
}

func NewModbusClientFromSettingsWithContext(ctx context.Context, logger *zap.Logger, settings *serial.ClientSettings) (client.ModbusClient, error) {
	config := settings.SerialSettings().ToPortConfig()
	port, err := sp.Open(config)
	if err != nil {
		return nil, err
	}
	transport := ascii.NewModbusTransport(port, logger)
	requestCreator := serial.NewSerialRequestCreator(ascii.NewModbusTransaction, ascii.NewModbusFrame)
	return client.NewModbusClient(ctx, logger, transport, requestCreator, settings.ResponseTimeout()), nil
}

func newModbusClient(logger *zap.Logger, stream io.ReadWriteCloser, responseTimeout time.Duration) client.ModbusClient {
	ctx := context.Background()
	transport := ascii.NewModbusTransport(stream, logger)
	requestCreator := serial.NewSerialRequestCreator(ascii.NewModbusTransaction, ascii.NewModbusFrame)
	return client.NewModbusClient(ctx, logger, transport, requestCreator, responseTimeout)
}
