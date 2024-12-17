package gomodbus

import (
	"bufio"
	"context"
	"errors"
	"io"
	"time"

	"github.com/goburrow/serial"
	"go.uber.org/zap"
)

func newRTUSerialPort(port serial.Port) *rtuSerialPort {
	return &rtuSerialPort{
		port:         port,
		lastActivity: time.Now(),
	}
}

type rtuSerialPort struct {
	port         serial.Port
	lastActivity time.Time
}

func (r *rtuSerialPort) Read(p []byte) (n int, err error) {
	b, e := r.port.Read(p)
	if e != nil {
		return b, e
	}
	r.lastActivity = time.Now()
	return b, e
}

func (r *rtuSerialPort) Write(p []byte) (n int, err error) {
	// we need to block for at least 3.5 character times between packets
	dwell := time.Since(r.lastActivity)
	if dwell < 1750*time.Microsecond {
		time.Sleep(1750*time.Microsecond - dwell)
	}
	b, e := r.port.Write(p)
	if e != nil {
		return b, e
	}
	r.lastActivity = time.Now()
	return b, e
}

func (r *rtuSerialPort) Close() error {
	return r.port.Close()
}

type ModbusRTUServer struct {
	modbusSerialServer
	running bool
}

func NewModbusRTUServer(logger *zap.Logger, port serial.Port, serverAddress uint16) (ModbusServer, error) {
	handler := NewDefaultHandler(logger, 65535, 65535, 65535, 65535)
	return NewModbusRTUServerWithHandler(logger, port, serverAddress, handler)
}

func NewModbusRTUServerWithHandler(logger *zap.Logger, port serial.Port, serverAddress uint16, handler RequestHandler) (ModbusServer, error) {
	if handler == nil {
		return nil, errors.New("handler is required")
	}
	internalPort := newRTUSerialPort(port)
	ctx, cancel := context.WithCancel(context.Background())
	return &ModbusRTUServer{
		modbusSerialServer: modbusSerialServer{
			port:    internalPort,
			address: serverAddress,
			reader:  bufio.NewReader(internalPort),
			modbusServer: modbusServer{
				cancelCtx: ctx,
				cancel:    cancel,
				logger:    logger,
				handler:   handler,
			},
		},
	}, nil
}

func newModbusRTUServerWithHandler(logger *zap.Logger, stream io.ReadWriteCloser, serverAddress uint16, handler RequestHandler) (*ModbusRTUServer, error) {
	if handler == nil {
		return nil, errors.New("handler is required")
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &ModbusRTUServer{
		modbusSerialServer: modbusSerialServer{
			address: serverAddress,
			port:    stream,
			reader:  bufio.NewReader(stream),
			modbusServer: modbusServer{
				cancelCtx: ctx,
				cancel:    cancel,
				logger:    logger,
				handler:   handler,
			},
		},
	}, nil
}

func formatRTUResponse(adu ApplicationDataUnit) []byte {
	data := adu.Bytes()
	return data
}

func (s *ModbusRTUServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return nil
	}

	s.logger.Info("Starting Modbus RTU server")
	go s.run()
	s.running = true
	return nil
}

func (s *ModbusRTUServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil && s.running {
		s.cancel()
	}

	s.logger.Info("Stopping Modbus RTU server")

	return nil
}

func (s *ModbusRTUServer) run() {
	s.logger.Debug("Starting Modbus RTU listener loop")
	s.logger.Debug("Flushing serial port until we find a packet")
	if flushedByteCount, err := s.flushPort(); err != nil {
		s.logger.Error("Failed to flush port", zap.Error(err))
		return
	} else {
		s.logger.Debug("Flushed port", zap.Int("bytesFlushed", flushedByteCount))
	}
	for {
		select {
		case <-s.cancelCtx.Done():
			return
		default:
			op, err := s.acceptAndValidateRequest()
			if err == errIgnorePacket {
				continue
			} else if err == io.EOF {
				s.logger.Warn("EOF received, did you set a timeout on the serial port? don't do that.")
				continue
			} else if err == ErrNotOurAddress {
				s.logger.Debug("Received request for different address")
				continue
			} else if err == ErrUnsupportedFunctionCode {
				// s.logger.Debug("Received request with unsupported function code, this is likely a parsing error")
				continue
			} else if err != nil {
				// s.logger.Error("Failed to accept request", zap.Error(err))
				continue
			}

			s.handlePacket(op)
		}
	}
}

func (s *ModbusRTUServer) flushPort() (int, error) {
	timeoutStart := time.Now()
	flushedByteCount := 0
	for {
		start := time.Now()
		_, _ = s.reader.ReadByte()
		readTime := time.Since(start)
		if readTime > 20*time.Millisecond {
			s.reader.UnreadByte()
			return flushedByteCount, nil
		}
		flushedByteCount++
		if time.Since(timeoutStart) > 5*time.Second {
			s.logger.Error("Failed to find packet start")
			return flushedByteCount, errTimeout
		}
	}
}

func (s *ModbusRTUServer) acceptAndValidateRequest() (ModbusOperation, error) {
	read := 0
	data := make([]byte, 256)
	d := make([]byte, 1)
	// We need, at a minimum, 2 bytes to read the address and function code, then we can read more
	for read < 2 {
		select {
		case <-s.cancelCtx.Done():
			return nil, s.cancelCtx.Err()
		default:
		}
		n, err := s.reader.Read(d)
		if err != nil {
			return nil, err
		}
		if n == 1 {
			data[read] = d[0]
			read++
		}
	}
	var err error
	var op ModbusOperation = nil
	functionCode := data[1]
	switch functionCode {
	case FunctionCodeReadCoils, FunctionCodeReadDiscreteInputs, FunctionCodeReadHoldingRegisters, FunctionCodeReadInputRegisters, FunctionCodeWriteSingleCoil, FunctionCodeWriteSingleRegister:
		// All of these functions are exactly 8 bytes long
		for read < 8 {
			select {
			case <-s.cancelCtx.Done():
				return nil, s.cancelCtx.Err()
			default:
			}
			n, err := s.reader.Read(d)
			if err != nil {
				return nil, err
			}
			if n == 1 {
				data[read] = d[0]
				read++
			}
		}
		op, err = NewModbusRTUOperation(data[:8], s.port, s.logger)
	case FunctionCodeWriteMultipleCoils, FunctionCodeWriteMultipleRegisters:
		// These functions have a variable length, so we need to read the length byte
		for read < 7 {
			select {
			case <-s.cancelCtx.Done():
				return nil, s.cancelCtx.Err()
			default:
			}
			n, err := s.reader.Read(d)
			if err != nil {
				return nil, err
			}
			if n == 1 {
				data[read] = d[0]
				read++
			}
		}
		byteCount := int(data[6])
		// 1 for address, 1 for function code, 2 for starting address, 2 for quantity, 1 for byte count, 2 for CRC which is 9 bytes
		// So we read the byteCount + 9 bytes
		for read < byteCount+9 {
			select {
			case <-s.cancelCtx.Done():
				return nil, s.cancelCtx.Err()
			default:
			}
			n, err := s.reader.Read(d)
			if err != nil {
				return nil, err
			}
			if n == 1 {
				data[read] = d[0]
				read++
			}
		}
		op, err = NewModbusRTUOperation(data[:byteCount+9], s.port, s.logger)
	default:
		return nil, ErrUnsupportedFunctionCode
	}

	if err != nil {
		return nil, err
	}

	if op.Address() != s.address {
		return nil, ErrNotOurAddress
	}

	return op, nil
}
