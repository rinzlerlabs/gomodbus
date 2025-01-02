package ascii

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/rinzlerlabs/gomodbus/common"
	"github.com/rinzlerlabs/gomodbus/data"
	"github.com/rinzlerlabs/gomodbus/transport"
	"github.com/rinzlerlabs/gomodbus/transport/serial"
	"go.uber.org/zap"
)

type modbusASCIITransport struct {
	logger          *zap.Logger
	mu              sync.Mutex
	stream          io.ReadWriteCloser
	reader          *bufio.Reader
	frameBuilder    transport.FrameBuilder
	responseTimeout time.Duration
	closing         bool
}

func NewModbusServerTransport(stream io.ReadWriteCloser, logger *zap.Logger) transport.Transport {
	return &modbusASCIITransport{
		logger:       logger,
		stream:       stream,
		reader:       bufio.NewReader(stream),
		frameBuilder: serial.NewFrameBuilder(NewModbusApplicationDataUnit),
	}
}

func NewModbusClientTransport(stream io.ReadWriteCloser, logger *zap.Logger, responseTimeout time.Duration) transport.Transport {
	return &modbusASCIITransport{
		logger:          logger,
		stream:          stream,
		reader:          bufio.NewReader(stream),
		frameBuilder:    serial.NewFrameBuilder(NewModbusApplicationDataUnit),
		responseTimeout: responseTimeout,
	}
}

func (t *modbusASCIITransport) readRawFrame(context.Context) ([]byte, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	str, err := t.reader.ReadString('\n')
	if err == io.EOF && t.closing {
		return nil, errors.Join(err, common.ErrTransportClosing)
	}
	if err != nil {
		return nil, err
	}
	return []byte(str), nil
}

func (t *modbusASCIITransport) ReadResponse(ctx context.Context, request transport.ApplicationDataUnit) (transport.ApplicationDataUnit, error) {
	bytesChan := make(chan []byte)
	errChan := make(chan error)

	go func() {
		bytes, err := t.readRawFrame(ctx)
		if err != nil {
			errChan <- err
			return
		}
		bytesChan <- bytes
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(t.responseTimeout):
		return nil, common.ErrTimeout
	case err := <-errChan:
		return nil, err
	case bytes := <-bytesChan:
		if op, ok := request.PDU().Operation().(data.CountableOperation); ok {
			return ParseModbusResponseFrame(bytes, op.Count())
		}
		return ParseModbusResponseFrame(bytes, 0)
	}
}

func (t *modbusASCIITransport) ReadRequest(ctx context.Context) (transport.ApplicationDataUnit, error) {
	bytes, err := t.readRawFrame(ctx)
	if err != nil {
		return nil, err
	}
	return ParseModbusRequestFrame(bytes)
}

func (t *modbusASCIITransport) WriteRequestFrame(address uint16, pdu *transport.ProtocolDataUnit) (transport.ApplicationDataUnit, error) {
	header := serial.NewHeader(address)
	adu, err := t.frameBuilder.BuildResponseFrame(header, pdu)
	if err != nil {
		return nil, err
	}

	_, err = t.Write([]byte(fmt.Sprintf(":%X\r\n", adu.Bytes())))
	if err != nil {
		return nil, err
	}
	return adu, nil
}

func (t *modbusASCIITransport) WriteResponseFrame(header transport.Header, pdu *transport.ProtocolDataUnit) error {
	adu, err := t.frameBuilder.BuildResponseFrame(header, pdu)
	if err != nil {
		return err
	}
	_, err = t.Write([]byte(fmt.Sprintf(":%X\r\n", adu.Bytes())))
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
	t.closing = true
	stream := t.stream
	t.stream = nil
	return stream.Close()
}
