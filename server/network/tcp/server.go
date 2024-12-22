package tcp

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"

	"github.com/rinzlerlabs/gomodbus/server"
	"github.com/rinzlerlabs/gomodbus/transport/network/tcp"
	"go.uber.org/zap"
)

func NewModbusServer(logger *zap.Logger, endpoint string) (server.ModbusServer, error) {
	handler := server.NewDefaultHandler(logger, 65535, 65535, 65535, 65535)
	return NewModbusServerWithHandler(logger, endpoint, handler)
}

func NewModbusServerWithHandler(logger *zap.Logger, endpoint string, handler server.RequestHandler) (server.ModbusServer, error) {
	if handler == nil {
		return nil, errors.New("handler is required")
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &modbusServer{
		logger:    logger,
		handler:   handler,
		cancelCtx: ctx,
		cancel:    cancel,
		endpoint:  endpoint,
	}, nil
}

type modbusServer struct {
	handler   server.RequestHandler
	cancelCtx context.Context
	cancel    context.CancelFunc
	logger    *zap.Logger
	isRunning bool
	endpoint  string
	wg        sync.WaitGroup
}

func (s *modbusServer) IsRunning() bool {
	return s.isRunning
}

func (s *modbusServer) Start() error {
	if s.isRunning {
		s.logger.Debug("Modbus TCP server already running")
		return nil
	}

	s.logger.Info("Starting Modbus TCP server")
	s.isRunning = true
	go s.run()
	return nil
}

func (s *modbusServer) Stop() error {
	s.logger.Info("Stopping Modbus TCP server")
	s.isRunning = false
	s.cancel()
	return nil
}

func (s *modbusServer) Close() error {
	s.logger.Info("Closing Modbus TCP server")
	s.cancel()
	s.logger.Info("Waiting for all clients to disconnect")
	s.wg.Wait()
	s.logger.Info("All clients disconnected")
	return nil
}

func (s *modbusServer) run() {
	listener, err := net.Listen("tcp", s.endpoint)
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

func (s *modbusServer) handleClient(conn net.Conn) {
	s.wg.Add(1)
	defer s.wg.Done()
	defer conn.Close()
	t := tcp.NewModbusTransport(conn, s.logger)
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
