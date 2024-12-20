package tcp

import (
	"context"
	"io"
	"net"
	"sync"

	"github.com/rinzlerlabs/gomodbus/transport"
	"go.uber.org/zap"
)

type readWriteCloseRemoteAddresser interface {
	io.ReadWriteCloser
	RemoteAddr() net.Addr
}

type modbusTCPSocketTransport struct {
	logger *zap.Logger
	mu     sync.Mutex
	conn   readWriteCloseRemoteAddresser
}

func (m *modbusTCPSocketTransport) readRawFrame(context.Context) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logger.Debug("Reading data from TCP socket")
	data := make([]byte, 260)
	n, err := m.conn.Read(data)
	if err != nil {
		return nil, err
	}
	data = data[:n]
	m.logger.Debug("Received data from TCP socket", zap.String("data", transport.EncodeToString(data)))
	return data, nil
}

// AcceptRequest implements transport.Transport.
func (m *modbusTCPSocketTransport) AcceptRequest(ctx context.Context) (transport.ModbusTransaction, error) {
	m.logger.Debug("Accepting request from TCP socket", zap.String("remoteAddr", m.conn.RemoteAddr().String()))
	dataChan := make(chan []byte)
	errChan := make(chan error)

	go func() {
		data, err := m.readRawFrame(ctx)
		if err != nil {
			errChan <- err
			return
		}

		dataChan <- data
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errChan:
		return nil, err
	case data := <-dataChan:
		frame, err := NewModbusRequestFrame(data)
		if err != nil {
			return nil, err
		}
		return NewModbusTransaction(frame, m), nil
	}
}

// Close implements transport.Transport.
func (m *modbusTCPSocketTransport) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logger.Debug("Closing TCP socket")
	return m.conn.Close()
}

// Flush implements transport.Transport.
func (m *modbusTCPSocketTransport) Flush(context.Context) error {
	m.logger.Debug("Flushing socket transport is a no-op")
	return nil
}

// Write implements transport.Transport.
func (m *modbusTCPSocketTransport) Write(p []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logger.Debug("Writing data to TCP socket", zap.String("data", transport.EncodeToString(p)))
	n, err := m.conn.Write(p)
	if err != nil {
		return 0, err
	}
	if n < len(p) {
		return n, io.ErrShortWrite
	}
	return n, nil
}

// WriteFrame implements transport.Transport.
func (m *modbusTCPSocketTransport) WriteFrame(frame *transport.ModbusFrame) error {
	_, err := m.Write(frame.Bytes())
	return err
}

func NewModbusTransport(conn readWriteCloseRemoteAddresser, logger *zap.Logger) transport.Transport {
	return &modbusTCPSocketTransport{
		logger: logger,
		conn:   conn,
	}
}
