package rtu

import (
	"bufio"
	"context"
	"io"
	"sync"
	"time"

	"github.com/rinzlerlabs/gomodbus/common"
	"github.com/rinzlerlabs/gomodbus/data"
	"github.com/rinzlerlabs/gomodbus/transport"
	. "github.com/rinzlerlabs/gomodbus/transport/serial"
	"go.uber.org/zap"
)

type modbusRTUTransport struct {
	logger *zap.Logger
	mu     sync.Mutex
	stream io.ReadWriteCloser
	reader *bufio.Reader
}

func NewModbusTransport(stream io.ReadWriteCloser, logger *zap.Logger) transport.Transport {
	return &modbusRTUTransport{
		logger: logger,
		stream: stream,
		reader: bufio.NewReader(stream),
	}
}

func (t *modbusRTUTransport) readRequestFrame(ctx context.Context) ([]byte, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	read := 0
	bytes := make([]byte, 256)
	d := make([]byte, 1)
	var err error
	// We need, at a minimum, 2 bytes to read the address and function code, then we can read more
	for read < 2 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		n, err := t.stream.Read(d)
		if err != nil {
			return nil, err
		}
		if n == 1 {
			bytes[read] = d[0]
			read++
		}
	}
	functionCode := data.FunctionCode(bytes[1])
	switch functionCode {
	case data.ReadCoils, data.ReadDiscreteInputs, data.ReadHoldingRegisters, data.ReadInputRegisters, data.WriteSingleCoil, data.WriteSingleRegister:
		// All of these functions are exactly 8 bytes long
		for read < 8 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
			n, err := t.stream.Read(d)
			if err != nil {
				return nil, err
			}
			if n == 1 {
				bytes[read] = d[0]
				read++
			}
		}
		return bytes[:8], nil
	case data.WriteMultipleCoils, data.WriteMultipleRegisters:
		// These functions have a variable length, so we need to read the length byte
		for read < 7 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
			n, err := t.stream.Read(d)
			if err != nil {
				return nil, err
			}
			if n == 1 {
				bytes[read] = d[0]
				read++
			}
		}
		byteCount := int(bytes[6])
		// 1 for address, 1 for function code, 2 for starting address, 2 for quantity, 1 for byte count, 2 for CRC which is 9 bytes
		// So we read the byteCount + 9 bytes
		for read < byteCount+9 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
			n, err := t.stream.Read(d)
			if err != nil {
				return nil, err
			}
			if n == 1 {
				if read == 256 {
					t.logger.Warn("Read too many bytes, discarding packet")
					return nil, ErrInvalidPacket
				}
				bytes[read] = d[0]
				read++
			}
		}
		return bytes[:byteCount+9], nil
	default:
		err = common.ErrUnsupportedFunctionCode
		return nil, err
	}
}

func (t *modbusRTUTransport) readResponseFrame(ctx context.Context) ([]byte, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	read := 0
	bytes := make([]byte, 256)
	d := make([]byte, 1)
	// We need, at a minimum, 2 bytes to read the address and function code, then we can read more
	for read < 2 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		n, err := t.stream.Read(d)
		if err != nil {
			return nil, err
		}
		if n == 1 {
			bytes[read] = d[0]
			read++
		}
	}
	functionCode := data.FunctionCode(bytes[1])
	switch functionCode {
	case data.ReadCoils, data.ReadDiscreteInputs, data.ReadHoldingRegisters, data.ReadInputRegisters:
		// These functions have a variable length, so we need to read the length byte
		// The length byte is the 3rd byte
		n, err := t.stream.Read(d)
		if err != nil {
			return nil, err
		}
		if n == 1 {
			bytes[read] = d[0]
			read++
		}
		length := int(bytes[2])
		// 3 for the bytes we already read, 2 for the CRC which is 5 bytes
		for read < length+5 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
			n, err := t.stream.Read(d)
			if err != nil {
				return nil, err
			}
			if n == 1 {
				bytes[read] = d[0]
				read++
			}
		}
	case data.WriteSingleCoil, data.WriteSingleRegister, data.WriteMultipleCoils, data.WriteMultipleRegisters:
		// These functions are exactly 8 bytes long
		for read < 8 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
			n, err := t.stream.Read(d)
			if err != nil {
				return nil, err
			}
			if n == 1 {
				bytes[read] = d[0]
				read++
			}
		}
	default:
		return nil, common.ErrUnsupportedFunctionCode
	}
	return bytes[:read], nil
}

func (t *modbusRTUTransport) AcceptRequest(ctx context.Context) (transport.ModbusTransaction, error) {
	data, err := t.readRequestFrame(ctx)
	if err != nil {
		return nil, err
	}

	frame, err := NewModbusRTURequestFrame(data)
	if err != nil {
		return nil, err
	}
	return NewModbusTransaction(frame, t), nil
}

func (t *modbusRTUTransport) WriteFrame(data *transport.ModbusFrame) error {
	_, err := t.Write(data.Bytes())
	return err
}

func (t *modbusRTUTransport) Write(p []byte) (int, error) {
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

func (t *modbusRTUTransport) Flush(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	timeoutStart := time.Now()
	flushedByteCount := 0
	for {
		start := time.Now()
		_, _ = t.reader.ReadByte()
		readTime := time.Since(start)
		if readTime > 20*time.Millisecond {
			t.reader.UnreadByte()
			t.logger.Debug("Flushed", zap.Int("bytesFlushed", flushedByteCount))
			return nil
		}
		flushedByteCount++
		if time.Since(timeoutStart) > 5*time.Second {
			t.logger.Error("Failed to find packet start")
			return common.ErrTimeout
		}
	}
}

func (t *modbusRTUTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	stream := t.stream
	t.stream = nil
	t.reader = nil
	if stream != nil {
		return stream.Close()
	}
	return nil
}
