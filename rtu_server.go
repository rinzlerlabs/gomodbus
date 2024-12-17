package gomodbus

import (
	"bufio"
	"context"
	"errors"
	"io"

	"github.com/tarm/serial"
	"go.uber.org/zap"
)

type ModbusRTUServer struct {
	modbusSerialServer
}

func NewModbusRTUServer(logger *zap.Logger, port *serial.Port, serverAddress uint16) (*ModbusRTUServer, error) {
	handler := NewDefaultHandler(logger, 65535, 65535, 65535, 65535)
	return NewModbusRTUServerWithHandler(logger, port, serverAddress, handler)
}

func NewModbusRTUServerWithHandler(logger *zap.Logger, port *serial.Port, serverAddress uint16, handler RequestHandler) (*ModbusRTUServer, error) {
	if handler == nil {
		return nil, errors.New("handler is required")
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &ModbusRTUServer{
		modbusSerialServer: modbusSerialServer{
			port:    port,
			address: serverAddress,
			reader:  bufio.NewReader(port),
			modbusServer: modbusServer{
				cancelCtx:           ctx,
				cancel:              cancel,
				logger:              logger,
				handler:             handler,
				responseCreatorFunc: NewRTUApplicationDataUnitFromResponse,
				responseFormatter:   formatRTUResponse,
				responseWriter:      port,
			},
		},
	}, nil
}

func newModbusRTUServerWithHandler(logger *zap.Logger, stream io.ReadWriter, serverAddress uint16, handler RequestHandler) (*ModbusRTUServer, error) {
	if handler == nil {
		return nil, errors.New("handler is required")
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &ModbusRTUServer{
		modbusSerialServer: modbusSerialServer{
			address: serverAddress,
			reader:  bufio.NewReader(stream),
			modbusServer: modbusServer{
				cancelCtx:           ctx,
				cancel:              cancel,
				logger:              logger,
				handler:             handler,
				responseCreatorFunc: NewRTUApplicationDataUnitFromResponse,
				responseFormatter:   formatRTUResponse,
				responseWriter:      stream,
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

	s.logger.Info("Starting Modbus RTU server")
	go s.run()
	return nil
}

func (s *ModbusRTUServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		s.cancel()
	}

	s.logger.Info("Stopping Modbus RTU server")

	return nil
}

func (s *ModbusRTUServer) run() {
	s.logger.Debug("Starting Modbus RTU listener loop")
	for {
		select {
		case <-s.cancelCtx.Done():
			return
		default:
			packet, err := s.acceptRequest()
			if err == io.EOF {
				s.logger.Warn("EOF received, did you set a timeout on the serial port? don't do that.")
				continue
			}
			if err != nil {
				s.logger.Error("Failed to accept request", zap.Error(err))
				continue
			}

			if packet.Address() != s.address {
				s.logger.Debug("Received packet with incorrect address, discarding packet", zap.Any("packet", packet))
				continue
			}
			s.handlePacket(packet)
		}
	}
}

func (s *ModbusRTUServer) acceptRequest() (ApplicationDataUnit, error) {
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
		return NewRTUApplicationDataUnitFromRequest(data[:8])
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
		return NewRTUApplicationDataUnitFromRequest(data[:byteCount+9])
	}
	return nil, nil
}
