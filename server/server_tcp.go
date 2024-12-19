package server

import (
	"context"
	"errors"
	"io"
	"net"

	"github.com/rinzlerlabs/gomodbus/transport/network/tcp"
	"go.uber.org/zap"
)

func NewModbusTCPServer(logger *zap.Logger, endpoint string) (ModbusServer, error) {
	handler := NewDefaultHandler(logger, 65535, 65535, 65535, 65535)
	return NewModbusTCPServerWithHandler(logger, endpoint, handler)
}

func NewModbusTCPServerWithHandler(logger *zap.Logger, endpoint string, handler RequestHandler) (ModbusServer, error) {
	if handler == nil {
		return nil, errors.New("handler is required")
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &modbusTCPServer{
		logger:    logger,
		handler:   handler,
		cancelCtx: ctx,
		cancel:    cancel,
		endpoint:  endpoint,
	}, nil
}

type modbusTCPServer struct {
	handler   RequestHandler
	cancelCtx context.Context
	cancel    context.CancelFunc
	logger    *zap.Logger
	isRunning bool
	endpoint  string
}

func (s *modbusTCPServer) IsRunning() bool {
	return s.isRunning
}

func (s *modbusTCPServer) Start() error {
	if s.isRunning {
		s.logger.Debug("Modbus TCP server already running")
		return nil
	}

	s.logger.Info("Starting Modbus TCP server")
	s.isRunning = true
	go s.run()
	return nil
}

func (s *modbusTCPServer) Stop() error {
	s.logger.Info("Stopping Modbus TCP server")
	s.isRunning = false
	s.cancel()
	return nil
}

func (s *modbusTCPServer) run() {
	listener, err := net.Listen("tcp", s.endpoint)
	if err != nil {
		s.logger.Error("Failed to listen", zap.Error(err))
		return
	}
	if err != nil {
		s.logger.Error("Failed to create TCP transport", zap.Error(err))
		return
	}
	defer listener.Close()

	for {
		select {
		case <-s.cancelCtx.Done():
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				s.logger.Error("Failed to accept connection", zap.Error(err))
				continue
			}
			go s.handleClient(conn)
		}
	}
}

func (s *modbusTCPServer) handleClient(conn net.Conn) {
	defer conn.Close()
	t := tcp.NewModbusTCPSocketTransport(conn, s.logger)
	for {
		select {
		case <-s.cancelCtx.Done():
			return
		default:
		}
		transaction, err := t.AcceptRequest(s.cancelCtx)
		if err == io.EOF {
			s.logger.Info("Client disconnected")
			err := t.Close()
			if err != nil {
				s.logger.Error("Failed to close connection", zap.Error(err))
			}
			return
		}
		if err != nil {
			s.logger.Error("Failed to accept request", zap.Error(err))
			return
		}
		s.handler.Handle(transaction)
	}
}
