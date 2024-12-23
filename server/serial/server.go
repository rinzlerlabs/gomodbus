package serial

import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/rinzlerlabs/gomodbus/common"
	"github.com/rinzlerlabs/gomodbus/server"
	"github.com/rinzlerlabs/gomodbus/transport"
	"go.uber.org/zap"
)

type ModbusSerialServer interface {
	server.ModbusServer
	Handler() server.RequestHandler
}

func NewModbusSerialServerWithHandler(logger *zap.Logger, serverAddress uint16, handler server.RequestHandler, transport transport.Transport) (ModbusSerialServer, error) {
	if handler == nil {
		return nil, errors.New("handler is required")
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &modbusSerialServer{
		logger:    logger,
		handler:   handler,
		cancelCtx: ctx,
		cancel:    cancel,
		address:   serverAddress,
		transport: transport,
		stats:     server.NewServerStats(),
	}, nil
}

type modbusSerialServer struct {
	handler   server.RequestHandler
	cancelCtx context.Context
	cancel    context.CancelFunc
	logger    *zap.Logger
	mu        sync.Mutex
	address   uint16
	transport transport.Transport
	isRunning bool
	wg        sync.WaitGroup
	stats     *server.ServerStats
}

func (s *modbusSerialServer) IsRunning() bool {
	return s.isRunning
}

func (s *modbusSerialServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.isRunning {
		s.logger.Debug("Modbus RTU server already running")
		return nil
	}

	if s.cancel == nil {
		ctx, cancel := context.WithCancel(context.Background())
		s.cancelCtx = ctx
		s.cancel = cancel
	}

	s.logger.Info("Starting Modbus RTU server")
	go s.run()
	return nil
}

func (s *modbusSerialServer) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logger.Info("Stopping Modbus server")

	if s.cancel != nil {
		s.cancel()
		defer func() { s.cancel = nil }()
	}
	s.wg.Wait()
	s.logger.Info("Modbus Server stopped")
	return s.transport.Close()
}

func (s *modbusSerialServer) Handler() server.RequestHandler {
	return s.handler
}

func (s *modbusSerialServer) Stats() *server.ServerStats {
	return s.stats
}

func (s *modbusSerialServer) run() {
	s.isRunning = true
	s.wg.Add(1)
	defer s.wg.Done()
	defer func() { s.isRunning = false }()

	s.logger.Debug("Starting Modbus RTU listener loop")
	s.logger.Debug("Flushing serial port until we find a packet")
	if err := s.transport.Flush(s.cancelCtx); err != nil {
		return
	}
	for {
		select {
		case <-s.cancelCtx.Done():
			return
		default:
			op, err := s.acceptAndValidateTransaction()
			if err == common.ErrIgnorePacket {
				continue
			} else if err == io.EOF {
				s.logger.Warn("EOF received, did you set a timeout on the serial port? don't do that.")
				continue
			} else if err == common.ErrNotOurAddress {
				s.logger.Debug("Received request for different address")
				continue
			} else if err == common.ErrUnsupportedFunctionCode {
				s.logger.Debug("Received request with unsupported function code, this is likely a timing error")
				continue
			} else if err != nil {
				// s.logger.Error("Failed to accept request", zap.Error(err))
				continue
			}
			s.stats.AddRequest(op)
			err = s.handler.Handle(op)
			if err != nil {
				s.stats.AddError(err)
				s.logger.Error("Failed to handle request", zap.Error(err))
			}
		}
	}
}

func (s *modbusSerialServer) acceptAndValidateTransaction() (transport.ModbusTransaction, error) {
	txn, err := s.transport.AcceptRequest(s.cancelCtx)
	if err != nil {
		return nil, err
	}

	if txn.Frame().Header().(transport.SerialHeader).Address() != s.address {
		return nil, common.ErrNotOurAddress
	}

	return txn, nil
}
