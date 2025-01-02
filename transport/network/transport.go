package network

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/rinzlerlabs/gomodbus/common"
	"github.com/rinzlerlabs/gomodbus/data"
	"github.com/rinzlerlabs/gomodbus/transport"
	"go.uber.org/zap"
)

type ReadWriteCloseRemoteAddresser interface {
	io.ReadWriteCloser
	RemoteAddr() net.Addr
}

type modbusTCPSocketTransport struct {
	logger          *zap.Logger
	mu              sync.Mutex
	conn            ReadWriteCloseRemoteAddresser
	frameBuilder    transport.FrameBuilder
	headerManager   *headerManager
	responseTimeout time.Duration
	closing         bool
	wg              sync.WaitGroup
}

func NewModbusServerTransport(conn ReadWriteCloseRemoteAddresser, logger *zap.Logger) transport.Transport {
	return &modbusTCPSocketTransport{
		logger:        logger,
		conn:          conn,
		frameBuilder:  NewFrameBuilder(),
		headerManager: &headerManager{},
	}
}

func NewModbusClientTransport(conn ReadWriteCloseRemoteAddresser, logger *zap.Logger, responseTimeout time.Duration) transport.Transport {
	return &modbusTCPSocketTransport{
		logger:          logger,
		conn:            conn,
		frameBuilder:    NewFrameBuilder(),
		headerManager:   &headerManager{},
		responseTimeout: responseTimeout,
	}
}

func (m *modbusTCPSocketTransport) readRawFrame(context.Context) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logger.Debug("Reading data from TCP socket")
	data := make([]byte, 260)
	n, err := m.conn.Read(data)
	if err == io.EOF && m.closing {
		return nil, errors.Join(err, common.ErrTransportClosing)
	}
	if err != nil {
		return nil, err
	}
	data = data[:n]
	m.logger.Debug("Received data from TCP socket", zap.String("data", common.EncodeToString(data)))
	return data, nil
}

func (m *modbusTCPSocketTransport) ReadRequest(ctx context.Context) (transport.ApplicationDataUnit, error) {
	m.logger.Debug("Accepting request from TCP socket", zap.String("remoteAddr", m.conn.RemoteAddr().String()))
	dataChan := make(chan transport.ApplicationDataUnit)
	errChan := make(chan error)
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
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
	case <-time.After(m.responseTimeout):
		return nil, common.ErrTimeout
	case err := <-errChan:
		return nil, err
	case data := <-dataChan:
		return data, nil
	}
}

func (m *modbusTCPSocketTransport) Close() error {
	m.logger.Debug("Closing TCP socket")
	m.closing = true
	err := m.conn.Close() // Doing this is going to cause errors to return from the read/write functions
	m.wg.Wait()
	return err
}

func (m *modbusTCPSocketTransport) Flush(context.Context) error {
	m.logger.Debug("Flushing socket transport is a no-op")
	return nil
}

func (m *modbusTCPSocketTransport) WriteRequestFrame(address uint16, pdu *transport.ProtocolDataUnit) (transport.ApplicationDataUnit, error) {
	header := m.headerManager.NewHeader()
	adu, err := m.frameBuilder.BuildResponseFrame(header, pdu)
	if err != nil {
		return nil, err
	}
	_, err = m.write(adu.Bytes())
	if err != nil {
		return nil, err
	}
	return adu, nil
}

func (m *modbusTCPSocketTransport) WriteResponseFrame(header transport.Header, pdu *transport.ProtocolDataUnit) error {
	adu, err := m.frameBuilder.BuildResponseFrame(header, pdu)
	if err != nil {
		return err
	}
	_, err = m.write(adu.Bytes())
	return err
}

func (m *modbusTCPSocketTransport) write(p []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logger.Debug("Writing data to TCP socket", zap.String("data", common.EncodeToString(p)))
	n, err := m.conn.Write(p)
	if err != nil {
		return 0, err
	}
	if n < len(p) {
		return n, io.ErrShortWrite
	}
	return n, nil
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

type headerManager struct {
	mu            sync.Mutex
	transactionID uint16
}

func (hm *headerManager) NewHeader() transport.Header {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.transactionID++
	txnId := []byte{byte(hm.transactionID >> 8), byte(hm.transactionID & 0xff)}
	return NewHeader(txnId, []byte{0x00, 0x00}, byte(0x01))
}
