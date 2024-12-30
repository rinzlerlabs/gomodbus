package network

import (
	"context"
	"io"
	"net"
	"sync"

	"github.com/rinzlerlabs/gomodbus/data"
	"github.com/rinzlerlabs/gomodbus/transport"
	"go.uber.org/zap"
)

type ReadWriteCloseRemoteAddresser interface {
	io.ReadWriteCloser
	RemoteAddr() net.Addr
}

type modbusTCPSocketTransport struct {
	logger *zap.Logger
	mu     sync.Mutex
	conn   ReadWriteCloseRemoteAddresser
}

func NewModbusServerTransport(conn ReadWriteCloseRemoteAddresser, logger *zap.Logger) transport.Transport {
	return &modbusTCPSocketTransport{
		logger: logger,
		conn:   conn,
	}
}

func NewModbusClientTransport(conn ReadWriteCloseRemoteAddresser, logger *zap.Logger) transport.Transport {
	return &modbusTCPSocketTransport{
		logger: logger,
		conn:   conn,
	}
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

func (m *modbusTCPSocketTransport) ReadRequest(ctx context.Context) (transport.ApplicationDataUnit, error) {
	m.logger.Debug("Accepting request from TCP socket", zap.String("remoteAddr", m.conn.RemoteAddr().String()))
	dataChan := make(chan transport.ApplicationDataUnit)
	errChan := make(chan error)

	go func() {
		data, err := m.readRequestFrame(ctx)
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
		return data, nil
	}
}

func (m *modbusTCPSocketTransport) ReadResponse(ctx context.Context, request transport.ApplicationDataUnit) (transport.ApplicationDataUnit, error) {
	dataChan := make(chan transport.ApplicationDataUnit)
	errChan := make(chan error)

	go func() {
		data, err := m.readResponseFrame(ctx, request)
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
		return data, nil
	}
}

func (m *modbusTCPSocketTransport) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logger.Debug("Closing TCP socket")
	return m.conn.Close()
}

func (m *modbusTCPSocketTransport) Flush(context.Context) error {
	m.logger.Debug("Flushing socket transport is a no-op")
	return nil
}

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

func (m *modbusTCPSocketTransport) WriteFrame(frame transport.ApplicationDataUnit) error {
	_, err := m.Write(frame.Bytes())
	return err
}

func (t *modbusTCPSocketTransport) readRequestFrame(ctx context.Context) (transport.ApplicationDataUnit, error) {
	data, err := t.readRawFrame(ctx)
	if err != nil {
		return nil, err
	}
	return ParseModbusRequestFrame(data)
}

func (t *modbusTCPSocketTransport) readResponseFrame(ctx context.Context, request transport.ApplicationDataUnit) (transport.ApplicationDataUnit, error) {
	bytes, err := t.readRawFrame(ctx)
	if err != nil {
		return nil, err
	}
	if op, ok := request.PDU().Operation().(data.CountableOperation); ok {
		return ParseModbusServerResponseFrame(bytes, op.Count())
	}
	return ParseModbusServerResponseFrame(bytes, 0)
}
