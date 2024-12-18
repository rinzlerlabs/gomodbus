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

func NewModbusASCIITransport(stream io.ReadWriteCloser, logger *zap.Logger) transport.Transport {
	return &modbusASCIITransport{
		logger: logger,
		stream: stream,
		reader: bufio.NewReader(stream),
	}
}

func NewModbusASCIIClientTransport(stream io.ReadWriteCloser, logger *zap.Logger) transport.Transport {
	return &modbusASCIITransport{
		logger: logger,
		stream: stream,
		reader: bufio.NewReader(stream),
	}
}

func (t *modbusASCIITransport) ReadNextRawFrame(ctx context.Context) ([]byte, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	str, err := t.reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	return []byte(str), nil
}

func (t *modbusASCIITransport) ReadNextFrame(ctx context.Context) (data.ModbusFrame, error) {
	str, err := t.ReadNextRawFrame(ctx)
	if err != nil {
		return nil, err
	}
	return NewModbusFrame(str, t)
}

func (t *modbusASCIITransport) WriteRawFrame(data []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	n, err := t.stream.Write(data)
	if err != nil {
		return err
	}
	if n < len(data) {
		return io.ErrShortWrite
	}
	return nil
}

func (t *modbusASCIITransport) WriteFrame(adu data.ModbusFrame) error {
	bytes := adu.Bytes()
	return t.WriteRawFrame([]byte(fmt.Sprintf(":%X\r\n", bytes)))
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
