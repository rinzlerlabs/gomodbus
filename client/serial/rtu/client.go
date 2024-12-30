package rtu

import (
	"context"
	"io"
	"net/url"
	"time"

	sp "github.com/goburrow/serial"
	"github.com/rinzlerlabs/gomodbus/client"
	"github.com/rinzlerlabs/gomodbus/client/serial"
	"github.com/rinzlerlabs/gomodbus/transport/serial/rtu"
	"go.uber.org/zap"
)

func NewModbusClientFromURI(logger *zap.Logger, uri string) (client.ModbusClient, error) {
	return NewModbusClientFromUriWithContext(context.Background(), logger, uri)
}

func NewModbusClientFromUriWithContext(ctx context.Context, logger *zap.Logger, uri string) (client.ModbusClient, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	settings, err := serial.NewClientSettingsFromURI(u)
	return NewModbusClientWithContext(ctx, logger, settings)
}

func NewModbusClient(logger *zap.Logger, settings *serial.ClientSettings) (client.ModbusClient, error) {
	return NewModbusClientWithContext(context.Background(), logger, settings)
}

func NewModbusClientWithContext(ctx context.Context, logger *zap.Logger, settings *serial.ClientSettings) (client.ModbusClient, error) {
	config := settings.SerialSettings().ToPortConfig()
	port, err := sp.Open(config)
	if err != nil {
		return nil, err
	}
	transport := rtu.NewModbusClientTransport(port, logger)
	requestCreator := serial.NewSerialRequestCreator(rtu.NewModbusTransaction, rtu.NewModbusFrame)
	return client.NewModbusClient(ctx, logger, transport, requestCreator, settings.ResponseTimeout()), nil
}

func newModbusClient(logger *zap.Logger, stream io.ReadWriteCloser, responseTimeout time.Duration) client.ModbusClient {
	ctx := context.Background()
	transport := rtu.NewModbusClientTransport(stream, logger)
	requestCreator := serial.NewSerialRequestCreator(rtu.NewModbusTransaction, rtu.NewModbusFrame)
	return client.NewModbusClient(ctx, logger, transport, requestCreator, responseTimeout)
}
