package rtu

import (
	"bufio"
	"context"
	"encoding/hex"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/rinzlerlabs/gomodbus/common"
	"github.com/rinzlerlabs/gomodbus/data"
	"github.com/rinzlerlabs/gomodbus/transport"
	"go.uber.org/zap"
)

type modbusRTUTransport struct {
	logger     *zap.Logger
	mu         sync.Mutex
	stream     io.ReadWriteCloser
	reader     *bufio.Reader
	serverAddr uint16
}

func NewModbusServerTransport(stream io.ReadWriteCloser, logger *zap.Logger, serverAddress uint16) transport.Transport {
	return &modbusRTUTransport{
		logger:     logger,
		stream:     stream,
		reader:     bufio.NewReader(stream),
		serverAddr: serverAddress,
	}
}

func NewModbusClientTransport(stream io.ReadWriteCloser, logger *zap.Logger) transport.Transport {
	return &modbusRTUTransport{
		logger: logger,
		stream: stream,
		reader: bufio.NewReader(stream),
	}
}

func (t *modbusRTUTransport) ReadRequest(ctx context.Context) (transport.ApplicationDataUnit, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
start:
	addressFound := false
	read := 0
	bytes := make([]byte, 256)
	header := make([]byte, 1)
	// We need, at a minimum, 2 bytes to read the address and function code, then we can read more
	for read < 2 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		n, err := t.stream.Read(header)
		if err != nil {
			return nil, err
		}
		// This is a bit of a cheat, basically, if the first byte we read isn't our address, it is almost certainly not the start of a packet
		// If this check fails, the default case on the function code switch will discard the packet
		if n == 1 && !addressFound {
			if header[0] != byte(t.serverAddr) {
				continue
			} else {
				addressFound = true
			}
		}
		copy(bytes[read:read+n], header[:n])
		read += n
	}
	functionCode := data.FunctionCode(bytes[1])
	switch functionCode {
	case data.ReadCoils, data.ReadDiscreteInputs, data.ReadHoldingRegisters, data.ReadInputRegisters, data.WriteSingleCoil, data.WriteSingleRegister:
		// All of these functions are exactly 8 bytes long
		for read < 8 {
			d := make([]byte, 8-read)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
			n, err := t.stream.Read(d)
			if err != nil {
				return nil, err
			}
			copy(bytes[read:read+n], d[:n])
			read += n
		}
		t.logger.Debug("WireFrame", zap.String("bytes", strings.ToUpper(hex.EncodeToString(bytes[:read]))))
	case data.WriteMultipleCoils, data.WriteMultipleRegisters:
		// These functions have a variable length, so we need to read the length byte
		for read < 7 {
			d := make([]byte, 7-read)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
			n, err := t.stream.Read(d)
			if err != nil {
				return nil, err
			}
			copy(bytes[read:read+n], d[:n])
			read += n
		}
		byteCount := int(bytes[6])
		if functionCode == data.WriteMultipleRegisters {
			// The byte count must be an even number
			if byteCount%2 != 0 {
				t.logger.Debug("Invalid byte count for WriteMultipleRegisters, this usually indicates a corrupt packet", zap.Int("byteCount", byteCount))
				goto start
			}
			registerCount := uint16(bytes[4])<<8 | uint16(bytes[5])
			// The byte count must be twice the register count
			if byteCount != int(registerCount*2) {
				t.logger.Debug("Invalid byte count for WriteMultipleRegisters, this usually indicates a corrupt packet", zap.Int("byteCount", byteCount))
				goto start
			}
		} else if functionCode == data.WriteMultipleCoils {
			registerCount := uint16(bytes[4])<<8 | uint16(bytes[5])
			if byteCount != int(registerCount/8) {
				t.logger.Debug("Invalid byte count for WriteMultipleCoils, this usually indicates a corrupt packet", zap.Int("byteCount", byteCount))
				goto start
			}
		}
		// 1 for address, 1 for function code, 2 for starting address, 2 for quantity, 1 for byte count, 2 for CRC which is 9 bytes
		// So we read the byteCount + 9 bytes
		bytesNeeded := byteCount + 9
		for read < bytesNeeded {
			d := make([]byte, bytesNeeded-read)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
			n, err := t.stream.Read(d)
			if err != nil {
				return nil, err
			}
			if read+n >= 256 {
				t.logger.Warn("Read too many bytes, discarding packet")
				return nil, common.ErrInvalidPacket
			}

			copy(bytes[read:read+n], d)
			read += n
		}
		t.logger.Debug("WireFrame", zap.String("bytes", strings.ToUpper(hex.EncodeToString(bytes[:read]))))
	default:
		// This likely means we have a timing error, so we discard the packet
		t.logger.Debug("Unsupported function code", zap.Uint8("functionCode", uint8(functionCode)))
		goto start
	}
	t.logger.Debug("WireFrame", zap.String("bytes", strings.ToUpper(hex.EncodeToString(bytes[:read]))))
	return ParseModbusRequestFrame(bytes[:read])
}

func (t *modbusRTUTransport) ReadResponse(ctx context.Context, request transport.ApplicationDataUnit) (transport.ApplicationDataUnit, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	read := 0
	bytes := make([]byte, 256)
	// We need, at a minimum, 2 bytes to read the address and function code, then we can read more
	for read < 2 {
		d := make([]byte, 2)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		n, err := t.stream.Read(d)
		if err != nil {
			return nil, err
		}
		copy(bytes[read:read+n], d[:n])
		read += n
	}
	functionCode := data.FunctionCode(bytes[1])
	switch functionCode {
	case data.ReadCoils, data.ReadDiscreteInputs, data.ReadHoldingRegisters, data.ReadInputRegisters:
		// These functions have a variable length, so we need to read the length byte
		// The length byte is the 3rd byte
		d := make([]byte, 1)
		n, err := t.stream.Read(d)
		if err != nil {
			return nil, err
		}
		if n == 0 {
			return nil, common.ErrInvalidPacket
		}
		copy(bytes[read:read+n], d[:n])
		read += n
		length := int(bytes[2])
		bytesNeeded := length + 5
		if bytesNeeded > 256 {
			t.logger.Warn("Request indicates it needs more than 256 bytes, this is likely a corrupt packet")
			return nil, common.ErrInvalidPacket
		}
		// 3 for the bytes we already read, 2 for the CRC which is 5 bytes
		for read < bytesNeeded {
			d := make([]byte, bytesNeeded-read)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
			n, err := t.stream.Read(d)
			if err != nil {
				return nil, err
			}
			if read+n >= 256 {
				t.logger.Warn("Read too many bytes, discarding packet")
				return nil, common.ErrInvalidPacket
			}

			copy(bytes[read:read+n], d)
			read += n
		}
	case data.WriteSingleCoil, data.WriteSingleRegister, data.WriteMultipleCoils, data.WriteMultipleRegisters:
		// These functions are exactly 8 bytes long
		for read < 8 {
			d := make([]byte, 8-read)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
			n, err := t.stream.Read(d)
			if err != nil {
				return nil, err
			}
			copy(bytes[read:read+n], d[:n])
			read += n
		}
	case data.ReadCoilsError, data.ReadDiscreteInputsError, data.ReadHoldingRegistersError, data.ReadInputRegistersError, data.WriteSingleCoilError, data.WriteSingleRegisterError, data.WriteMultipleCoilsError, data.WriteMultipleRegistersError:
		// These functions are exactly 5 bytes long
		for read < 5 {
			d := make([]byte, 5-read)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
			n, err := t.stream.Read(d)
			if err != nil {
				return nil, err
			}
			copy(bytes[read:read+n], d[:n])
			read += n
		}
	default:
		return nil, common.ErrUnsupportedFunctionCode
	}
	t.logger.Debug("WireFrame", zap.String("bytes", strings.ToUpper(hex.EncodeToString(bytes[:read]))))

	if op, ok := request.PDU().Operation().(data.CountableOperation); ok {
		return ParseModbusResponseFrame(bytes[:read], op.Count())
	}
	return ParseModbusResponseFrame(bytes[:read], 0)
}

func (t *modbusRTUTransport) WriteFrame(data transport.ApplicationDataUnit) error {
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
