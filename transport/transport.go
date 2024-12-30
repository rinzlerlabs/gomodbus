package transport

import (
	"context"

	"github.com/rinzlerlabs/gomodbus/common"
)

type Transport interface {
	Flush(context.Context) error
	ReadRequest(context.Context) (ApplicationDataUnit, error)
	ReadResponse(context.Context, ApplicationDataUnit) (ApplicationDataUnit, error)
	WriteFrame(ApplicationDataUnit) error
	Write([]byte) (int, error)
	Close() error
}

type NilTransport struct{}

func (t *NilTransport) Flush(context.Context) error {
	return common.ErrNotImplemented
}

func (t *NilTransport) ReadRequest(context.Context) (ApplicationDataUnit, error) {
	return nil, common.ErrNotImplemented
}

func (t *NilTransport) ReadResponse(context.Context) (ApplicationDataUnit, error) {
	return nil, common.ErrNotImplemented
}

func (t *NilTransport) Write([]byte) error {
	return common.ErrNotImplemented
}

func (t *NilTransport) WriteFrame(ApplicationDataUnit) error {
	return common.ErrNotImplemented
}

func (t *NilTransport) Close() error {
	return common.ErrNotImplemented
}
