package ascii

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/rinzlerlabs/gomodbus/data"
	"github.com/rinzlerlabs/gomodbus/transport"
	"go.uber.org/zap"
)

type modbusASCIITransport struct {
	logger *zap.Logger
	mu     sync.Mutex
	stream io.ReadWriteCloser
	reader *bufio.Reader
}

func NewModbusServerTransport(stream io.ReadWriteCloser, logger *zap.Logger) transport.Transport {
	return &modbusASCIITransport{
		logger: logger,
		stream: stream,
		reader: bufio.NewReader(stream),
	}
}

func NewModbusClientTransport(stream io.ReadWriteCloser, logger *zap.Logger) transport.Transport {
	return &modbusASCIITransport{
		logger: logger,
		stream: stream,
		reader: bufio.NewReader(stream),
	}
}

func (t *modbusASCIITransport) readRawFrame(context.Context) ([]byte, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	str, err := t.reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	return []byte(str), nil
}

func (t *modbusASCIITransport) ReadResponse(ctx context.Context, request transport.ApplicationDataUnit) (transport.ApplicationDataUnit, error) {
	bytes, err := t.readRawFrame(ctx)
	if err != nil {
		return nil, err
	}
	if op, ok := request.PDU().Operation().(data.CountableOperation); ok {
		return ParseModbusResponseFrame(bytes, op.Count())
	}
	return ParseModbusResponseFrame(bytes, 0)
}

func (t *modbusASCIITransport) ReadRequest(ctx context.Context) (transport.ApplicationDataUnit, error) {
	bytes, err := t.readRawFrame(ctx)
	if err != nil {
		return nil, err
	}
	return ParseModbusRequestFrame(bytes)
}

func (t *modbusASCIITransport) WriteFrame(frame transport.ApplicationDataUnit) error {
	_, err := t.Write([]byte(fmt.Sprintf(":%X\r\n", frame.Bytes())))
	return err
}

func (t *modbusASCIITransport) Write(p []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	n, err := t.stream.Write(p)
	if err != nil {
		return 0, err
	}
	if n < len(p) {
		return n, io.ErrShortWrite
	}
	return n, nil
}

func (t *modbusASCIITransport) Flush(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.logger.Debug("Flushing ascii transport is a no-op")
	return nil
}

func (t *modbusASCIITransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	stream := t.stream
	t.stream = nil
	return stream.Close()
}
