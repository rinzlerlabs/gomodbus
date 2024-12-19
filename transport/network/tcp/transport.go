package tcp

import (
	"context"
	"io"
	"net"
	"sync"

	"github.com/rinzlerlabs/gomodbus/common"
	"github.com/rinzlerlabs/gomodbus/transport"
	"go.uber.org/zap"
)

type modbusTCPTransport struct {
	logger   *zap.Logger
	mu       sync.Mutex
	listener net.Listener
}

func NewModbusTCPTransport(endpoint string, logger *zap.Logger) (transport.Transport, error) {
	listener, err := net.Listen("tcp", endpoint)
	if err != nil {
		logger.Error("Failed to listen", zap.Error(err))
		return nil, err
	}

	return &modbusTCPTransport{
		logger:   logger,
		listener: listener,
	}, nil
}

func (t *modbusTCPTransport) AcceptRequest(ctx context.Context) (transport.ModbusTransaction, error) {
	connChan := make(chan net.Conn)
	errChan := make(chan error)

	go func() {
		conn, err := t.listener.Accept()
		if err != nil {
			errChan <- err
			return
		}
		t.logger.Debug("Accepted connection from TCP socket", zap.String("remoteAddr", conn.RemoteAddr().String()))
		connChan <- conn
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errChan:
		return nil, err
	case conn := <-connChan:
		socketTransport := newModbusTCPSocketTransport(conn, t.logger)
		return socketTransport.AcceptRequest(ctx)
	}
}

func (t *modbusTCPTransport) WriteRawFrame(data []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return common.ErrNotImplemented
}

func (t *modbusTCPTransport) WriteFrame(frame *transport.ModbusFrame) error {
	return common.ErrNotImplemented
}

func (t *modbusTCPTransport) Write(p []byte) (int, error) {
	return 0, common.ErrNotImplemented
}

func (t *modbusTCPTransport) Flush(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.logger.Debug("Flushing TCP transport is a no-op")
	return nil
}

func (t *modbusTCPTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	listener := t.listener
	t.listener = nil
	if listener == nil {
		return nil
	}
	return listener.Close()
}

type modbusTCPSocketTransport struct {
	logger *zap.Logger
	mu     sync.Mutex
	conn   net.Conn
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

func newModbusTCPSocketTransport(conn net.Conn, logger *zap.Logger) transport.Transport {
	return &modbusTCPSocketTransport{
		logger: logger,
		conn:   conn,
	}
}
