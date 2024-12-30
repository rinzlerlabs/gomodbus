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
	"github.com/rinzlerlabs/gomodbus/transport/serial"
	"go.uber.org/zap"
)

type modbusRTUTransport struct {
	logger          *zap.Logger
	mu              sync.Mutex
	frameBuilder    transport.FrameBuilder
	stream          io.ReadWriteCloser
	reader          *bufio.Reader
	serverAddr      uint16
	responseTimeout time.Duration
}

func NewModbusServerTransport(stream io.ReadWriteCloser, logger *zap.Logger, serverAddress uint16) transport.Transport {
	return &modbusRTUTransport{
		logger:          logger,
		stream:          stream,
		frameBuilder:    serial.NewFrameBuilder(NewModbusApplicationDataUnit),
		reader:          bufio.NewReader(stream),
		serverAddr:      serverAddress,
		responseTimeout: 5 * time.Second,
	}
}

func NewModbusClientTransport(stream io.ReadWriteCloser, logger *zap.Logger, responseTimeout time.Duration) transport.Transport {
	return &modbusRTUTransport{
		logger:          logger,
		stream:          stream,
		frameBuilder:    serial.NewFrameBuilder(NewModbusApplicationDataUnit),
		reader:          bufio.NewReader(stream),
		responseTimeout: responseTimeout,
	}
}

func (t *modbusRTUTransport) readWithTimeout(ctx context.Context, timeout time.Duration, bytes []byte, pos int) (int, error) {
	dataChan := make(chan int)
	errChan := make(chan error)

	go func() {
		read := 0
		for read < len(bytes) {
			d := make([]byte, len(bytes)-read)
			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			default:
				n, err := t.stream.Read(d)
				if err != nil {
					errChan <- err
					return
				}
				copy(bytes[read:read+n], d[:n])
				read += n
				if read == len(bytes) {
					dataChan <- read
				}
			}
		}
	}()
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case <-time.After(timeout):
		return 0, common.ErrTimeout
	case err := <-errChan:
		return 0, err
	case read := <-dataChan:
		return read + pos, nil
	}
}

func (t *modbusRTUTransport) ReadRequest(ctx context.Context) (transport.ApplicationDataUnit, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
start:
	read := 0
	bytes := make([]byte, 256)
	// We need, at a minimum, 2 bytes to read the address and function code, then we can read more
	read, err := t.readWithTimeout(ctx, t.responseTimeout, bytes[read:read+2], read)
	if err != nil {
		t.logger.Debug("Failed to read header bytes", zap.Error(err))
		return nil, err
	}
	// This is a bit of a cheat, basically, if the first byte we read isn't our address, it is almost certainly not the start of a packet
	// If this check fails, the default case on the function code switch will discard the packet
	if bytes[0] != byte(t.serverAddr) {
		goto start
	}

	functionCode := data.FunctionCode(bytes[1])
	switch functionCode {
	case data.ReadCoils, data.ReadDiscreteInputs, data.ReadHoldingRegisters, data.ReadInputRegisters, data.WriteSingleCoil, data.WriteSingleRegister:
		// All of these functions are exactly 8 bytes long
		read, err = t.readWithTimeout(ctx, t.responseTimeout, bytes[read:8], read)
		if err != nil {
			t.logger.Debug("Failed to read body bytes", zap.Error(err))
			return nil, err
		}
		t.logger.Debug("WireFrame", zap.String("bytes", strings.ToUpper(hex.EncodeToString(bytes[:read]))))
	case data.WriteMultipleCoils, data.WriteMultipleRegisters:
		// These functions have a variable length, so we need to read the length byte
		read, err = t.readWithTimeout(ctx, t.responseTimeout, bytes[read:7], read)
		if err != nil {
			t.logger.Debug("Failed to read body bytes", zap.Error(err))
			return nil, err
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
		read, err = t.readWithTimeout(ctx, t.responseTimeout, bytes[read:bytesNeeded], read)
		if err != nil {
			t.logger.Debug("Failed to read body bytes", zap.Error(err))
			return nil, err
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
	read, err := t.readWithTimeout(ctx, t.responseTimeout, bytes[read:read+2], read)
	if err != nil {
		return nil, err
	}
	functionCode := data.FunctionCode(bytes[1])
	switch functionCode {
	case data.ReadCoils, data.ReadDiscreteInputs, data.ReadHoldingRegisters, data.ReadInputRegisters:
		// These functions have a variable length, so we need to read the length byte
		// The length byte is the 3rd byte
		read, err = t.readWithTimeout(ctx, t.responseTimeout, bytes[read:read+1], read)
		if err != nil {
			return nil, err
		}

		length := int(bytes[2])
		bytesNeeded := length + 5
		if bytesNeeded > 256 {
			t.logger.Warn("Request indicates it needs more than 256 bytes, this is likely a corrupt packet")
			return nil, common.ErrInvalidPacket
		}
		// 3 for the bytes we already read, 2 for the CRC which is 5 bytes
		read, err = t.readWithTimeout(ctx, t.responseTimeout, bytes[read:bytesNeeded], read)
		if err != nil {
			return nil, err
		}
	case data.WriteSingleCoil, data.WriteSingleRegister, data.WriteMultipleCoils, data.WriteMultipleRegisters:
		// These functions are exactly 8 bytes long
		read, err = t.readWithTimeout(ctx, t.responseTimeout, bytes[read:8], read)
		if err != nil {
			return nil, err
		}
	case data.ReadCoilsError, data.ReadDiscreteInputsError, data.ReadHoldingRegistersError, data.ReadInputRegistersError, data.WriteSingleCoilError, data.WriteSingleRegisterError, data.WriteMultipleCoilsError, data.WriteMultipleRegistersError:
		// These functions are exactly 5 bytes long
		read, err = t.readWithTimeout(ctx, t.responseTimeout, bytes[read:5], read)
		if err != nil {
			return nil, err
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

func (t *modbusRTUTransport) WriteRequestFrame(address uint16, pdu *transport.ProtocolDataUnit) (transport.ApplicationDataUnit, error) {
	header := serial.NewHeader(address)
	adu, err := t.frameBuilder.BuildResponseFrame(header, pdu)
	if err != nil {
		return nil, err
	}
	_, err = t.write(adu.Bytes())
	if err != nil {
		return nil, err
	}
	return adu, nil
}

func (t *modbusRTUTransport) WriteResponseFrame(header transport.Header, pdu *transport.ProtocolDataUnit) error {
	adu, err := t.frameBuilder.BuildResponseFrame(header, pdu)
	if err != nil {
		return err
	}
	_, err = t.write(adu.Bytes())
	return err
}

func (t *modbusRTUTransport) write(p []byte) (int, error) {
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
		if time.Since(timeoutStart) > t.responseTimeout {
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
