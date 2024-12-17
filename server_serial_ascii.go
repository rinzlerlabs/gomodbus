package gomodbus

import (
	"bufio"
	"context"
	"errors"
	"io"

	"github.com/goburrow/serial"
	"go.uber.org/zap"
)

type ModbusASCIIServer struct {
	modbusSerialServer
}

func NewModbusASCIIServer(logger *zap.Logger, port serial.Port, serverAddress uint16) (ModbusServer, error) {
	handler := NewDefaultHandler(logger, 65535, 65535, 65535, 65535)
	return NewModbusASCIIServerWithHandler(logger, port, serverAddress, handler)
}

func NewModbusASCIIServerWithHandler(logger *zap.Logger, port serial.Port, serverAddress uint16, handler RequestHandler) (ModbusServer, error) {
	if handler == nil {
		return nil, errors.New("handler is required")
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &ModbusASCIIServer{
		modbusSerialServer: modbusSerialServer{
			address: serverAddress,
			reader:  bufio.NewReader(port),
			modbusServer: modbusServer{
				cancelCtx: ctx,
				cancel:    cancel,
				logger:    logger,
				handler:   handler,
			},
		},
	}, nil
}

// newModbusASCIIServerWithHandler creates a new Modbus ASCII server with a io.ReadWriter stream instead of an explicit port, for testing purposes, and a RequestHandler.
func newModbusASCIIServerWithHandler(logger *zap.Logger, stream io.ReadWriteCloser, serverAddress uint16, handler RequestHandler) (*ModbusASCIIServer, error) {
	if handler == nil {
		return nil, errors.New("handler is required")
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &ModbusASCIIServer{
		modbusSerialServer: modbusSerialServer{
			address: serverAddress,
			reader:  bufio.NewReader(stream),
			port:    stream,
			modbusServer: modbusServer{
				cancelCtx: ctx,
				cancel:    cancel,
				logger:    logger,
				handler:   handler,
			},
		},
	}, nil
}

func (s *ModbusASCIIServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info("Starting Modbus ASCII server")
	go s.run()

	return nil
}

func (s *ModbusASCIIServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		s.cancel()
	}

	s.logger.Info("Stopping Modbus ASCII server")

	return nil
}

func validateLRCChecksum(wireChecksum uint16, adu ApplicationDataUnit) error {
	lrc := adu.Checksum()
	if len(lrc) != 1 {
		return ErrInvalidChecksum
	}
	if lrc[0] != byte(wireChecksum) {
		return ErrInvalidChecksum
	}
	return nil
}

func (s *ModbusASCIIServer) run() {
	for {
		select {
		case <-s.cancelCtx.Done():
			return
		default:
			op, err := s.acceptAndValidateRequest()
			if err == ErrNotOurAddress {
				s.logger.Debug("Received packet with incorrect address, discarding packet", zap.Any("packet", op))
				continue
			} else if err != nil {
				s.logger.Error("Failed to accept request", zap.Error(err))
				continue
			}

			s.handlePacket(op)
		}
	}
}

func (s *ModbusASCIIServer) acceptAndValidateRequest() (ModbusOperation, error) {
	str, err := s.reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	op, err := NewModbusASCIIOperation(str, s.port, s.logger)
	if err != nil {
		return nil, err
	}

	if op.Address() != s.address {
		return nil, ErrNotOurAddress
	}

	return op, nil
}
