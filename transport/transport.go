package transport

import (
	"context"

	"github.com/rinzlerlabs/gomodbus/data"
)

type Transport interface {
	Flush(context.Context) error
	ReadNextRawFrame(context.Context) ([]byte, error)
	ReadNextFrame(context.Context) (data.ModbusFrame, error)
	WriteRawFrame([]byte) error
	WriteFrame(data.ModbusFrame) error
	Close() error
}
