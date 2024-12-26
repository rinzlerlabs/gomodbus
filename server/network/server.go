package network

import (
	"context"
	"errors"
	"io"
	"net"
	"net/url"
	"sync"

	"github.com/rinzlerlabs/gomodbus/server"
	"github.com/rinzlerlabs/gomodbus/transport/network"
	"go.uber.org/zap"
)

func NewModbusServer(logger *zap.Logger, endpoint string) (server.ModbusServer, error) {
	handler := server.NewDefaultHandler(logger, server.DefaultCoilCount, server.DefaultDiscreteInputCount, server.DefaultHoldingRegisterCount, server.DefaultInputRegisterCount)
	return NewModbusServerWithHandler(logger, endpoint, handler)
}

func NewModbusServerWithHandler(logger *zap.Logger, endpoint string, handler server.RequestHandler) (server.ModbusServer, error) {
	if handler == nil {
		return nil, errors.New("handler is required")
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	listener, err := net.Listen(u.Scheme, u.Host)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())

	return &modbusServer{
		logger:    logger,
		handler:   handler,
		cancelCtx: ctx,
		cancel:    cancel,
		listener:  listener,
		stats:     server.NewServerStats(),
	}, nil
}

func newModbusServerWithHandler(logger *zap.Logger, listener net.Listener, handler server.RequestHandler) (server.ModbusServer, error) {
	if handler == nil {
		return nil, errors.New("handler is required")
	}
	ctx, cancel := context.WithCancel(context.Background())

	return &modbusServer{
		logger:    logger,
		handler:   handler,
		cancelCtx: ctx,
		cancel:    cancel,
		listener:  listener,
		stats:     server.NewServerStats(),
	}, nil
}

type modbusServer struct {
	handler   server.RequestHandler
	cancelCtx context.Context
	cancel    context.CancelFunc
	logger    *zap.Logger
	isRunning bool
	listener  net.Listener
	stats     *server.ServerStats
	wg        sync.WaitGroup
	mu        sync.Mutex
}

func (s *modbusServer) IsRunning() bool {
	return s.isRunning
}

func (s *modbusServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
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
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logger.Info("Stopping Modbus TCP server")
	s.isRunning = false
	s.cancel()
	s.wg.Wait()
	return nil
}

func (s *modbusServer) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logger.Info("Closing Modbus TCP server")
	err := s.listener.Close()
	s.cancel()
	s.logger.Info("Waiting for all clients to disconnect")
	s.wg.Wait()
	s.logger.Info("All clients disconnected")
	return err
}

func (s *modbusServer) Stats() *server.ServerStats {
	return s.stats
}

func (s *modbusServer) run() {
	for {
		select {
		case <-s.cancelCtx.Done():
			return
		default:
			conn, err := s.listener.Accept()
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
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
	t := network.NewModbusTransport(conn, s.logger)
	defer t.Close()
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
		s.stats.AddRequest(transaction)
		err = s.handler.Handle(transaction)
		if err != nil {
			s.stats.AddError(err)
			s.logger.Error("Failed to handle request", zap.Error(err))
		}
	}
}
