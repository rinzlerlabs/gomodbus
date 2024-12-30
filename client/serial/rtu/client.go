package rtu

import (
	"context"
	"net/url"

	sp "github.com/goburrow/serial"
	"github.com/rinzlerlabs/gomodbus/client"
	"github.com/rinzlerlabs/gomodbus/client/serial"
	"github.com/rinzlerlabs/gomodbus/transport"
	st "github.com/rinzlerlabs/gomodbus/transport/serial"
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
	if err != nil {
		return nil, err
	}
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
	t := rtu.NewModbusClientTransport(port, logger)
	newHeader := func(address uint16) transport.Header {
		return st.NewHeader(address)
	}
	requestCreator := serial.NewSerialRequestCreator(newHeader, rtu.NewModbusRequest)
	return client.NewModbusClient(ctx, logger, t, requestCreator, settings.ResponseTimeout()), nil
}
