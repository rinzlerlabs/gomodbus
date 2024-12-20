package transport

import (
	"context"

	"github.com/rinzlerlabs/gomodbus/common"
)

type Transport interface {
	Flush(context.Context) error
	AcceptRequest(context.Context) (ModbusTransaction, error)
	WriteFrame(*ModbusFrame) error
	Write([]byte) (int, error)
	Close() error
}

type NilTransport struct{}

func (t *NilTransport) Flush(context.Context) error {
	return common.ErrNotImplemented
}

func (t *NilTransport) ReadNextRawFrame(context.Context) ([]byte, error) {
	return nil, common.ErrNotImplemented
}

func (t *NilTransport) ReadNextFrame(context.Context) (*ModbusFrame, error) {
	return nil, common.ErrNotImplemented
}

func (t *NilTransport) WriteRawFrame([]byte) error {
	return common.ErrNotImplemented
}

func (t *NilTransport) WriteFrame(*ModbusFrame) error {
	return common.ErrNotImplemented
}

func (t *NilTransport) Close() error {
	return common.ErrNotImplemented
}
