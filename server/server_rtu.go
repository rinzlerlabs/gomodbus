package server

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/goburrow/serial"
	"github.com/rinzlerlabs/gomodbus/transport/serial/rtu"
	"go.uber.org/zap"
)

func NewModbusRTUServer(logger *zap.Logger, port serial.Port, serverAddress uint16) (ModbusServer, error) {
	handler := NewDefaultHandler(logger, 65535, 65535, 65535, 65535)
	return NewModbusRTUServerWithHandler(logger, port, serverAddress, handler)
}

func NewModbusRTUServerWithHandler(logger *zap.Logger, port serial.Port, serverAddress uint16, handler RequestHandler) (ModbusServer, error) {
	if handler == nil {
		return nil, errors.New("handler is required")
	}
	internalPort := newRTUSerialPort(port)
	ctx, cancel := context.WithCancel(context.Background())
	return &modbusServer{
		logger:    logger,
		handler:   handler,
		cancelCtx: ctx,
		cancel:    cancel,
		address:   serverAddress,
		transport: rtu.NewModbusRTUServerTransport(internalPort, logger),
	}, nil
}

func newModbusRTUServerWithHandler(logger *zap.Logger, stream io.ReadWriteCloser, serverAddress uint16, handler RequestHandler) (*modbusServer, error) {
	if handler == nil {
		return nil, errors.New("handler is required")
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &modbusServer{
		logger:    logger,
		handler:   handler,
		cancelCtx: ctx,
		cancel:    cancel,
		address:   serverAddress,
		transport: rtu.NewModbusRTUServerTransport(stream, logger),
	}, nil
}

func newRTUSerialPort(port serial.Port) *rtuSerialPort {
	return &rtuSerialPort{
		port:         port,
		lastActivity: time.Now(),
	}
}

type rtuSerialPort struct {
	io.ReadWriteCloser
	port         serial.Port
	lastActivity time.Time
}

func (r *rtuSerialPort) Read(p []byte) (n int, err error) {
	b, e := r.port.Read(p)
	if e != nil {
		return b, e
	}
	r.lastActivity = time.Now()
	return b, e
}

func (r *rtuSerialPort) Write(p []byte) (n int, err error) {
	// we need to block for at least 3.5 character times between packets
	dwell := time.Since(r.lastActivity)
	if dwell < 1750*time.Microsecond {
		time.Sleep(1750*time.Microsecond - dwell)
	}
	b, e := r.port.Write(p)
	if e != nil {
		return b, e
	}
	r.lastActivity = time.Now()
	return b, e
}

func (r *rtuSerialPort) Close() error {
	return r.port.Close()
}
