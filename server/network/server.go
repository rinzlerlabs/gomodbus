package network

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"

	"github.com/rinzlerlabs/gomodbus/server"
	settings "github.com/rinzlerlabs/gomodbus/settings/network"
	"github.com/rinzlerlabs/gomodbus/transport"
	"github.com/rinzlerlabs/gomodbus/transport/network"
	"go.uber.org/zap"
)

func NewModbusServer(logger *zap.Logger, uri string) (server.ModbusServer, error) {
	settings, err := settings.NewServerSettingsFromURI(uri)
	if err != nil {
		return nil, err
	}
	return NewModbusServerFromSettings(logger, settings)
}

func NewModbusServerFromSettings(logger *zap.Logger, serverSettings *settings.ServerSettings) (server.ModbusServer, error) {
	handler := server.NewDefaultHandler(logger, server.DefaultCoilCount, server.DefaultDiscreteInputCount, server.DefaultHoldingRegisterCount, server.DefaultInputRegisterCount)
	return NewModbusServerWithHandler(logger, serverSettings, handler)
}

func NewModbusServerWithHandler(logger *zap.Logger, serverSettings *settings.ServerSettings, handler server.RequestHandler) (server.ModbusServer, error) {
	if handler == nil {
		return nil, errors.New("handler is required")
	}
	ctx, cancel := context.WithCancel(context.Background())

	return &modbusServer{
		logger:       logger,
		handler:      handler,
		cancelCtx:    ctx,
		cancel:       cancel,
		stats:        server.NewServerStats(),
		frameBuilder: nil,
		settings:     serverSettings,
	}, nil
}

type modbusServer struct {
	handler      server.RequestHandler
	cancelCtx    context.Context
	cancel       context.CancelFunc
	logger       *zap.Logger
	isRunning    bool
	settings     *settings.ServerSettings
	listener     net.Listener
	frameBuilder transport.FrameBuilder
	stats        *server.ServerStats
	wg           sync.WaitGroup
	mu           sync.Mutex
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

	if s.handler == nil {
		s.logger.Error("Handler is required")
		return errors.New("handler is required")
	}

	if s.listener == nil {
		listener, err := net.Listen(s.settings.Endpoint.Scheme, s.settings.Endpoint.Host)
		if err != nil {
			s.logger.Error("Failed to listen", zap.Error(err))
			return err
		}
		s.listener = listener
	}

	s.logger.Info("Starting Modbus TCP server")
	s.isRunning = true
	go s.run()
	return nil
}

func (s *modbusServer) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logger.Info("Closing Modbus TCP server")
	err := s.listener.Close()
	if err != nil {
		s.logger.Error("Error closing listener", zap.Error(err))
	}
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
	t := network.NewModbusServerTransport(conn, s.logger)
	defer t.Close()
	for {
		select {
		case <-s.cancelCtx.Done():
			return
		default:
		}
		op, err := t.ReadRequest(s.cancelCtx)
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
		s.stats.AddRequest(op)
		resp, err := s.handler.Handle(op)
		if err != nil {
			s.stats.AddError(err)
			s.logger.Error("Failed to handle request", zap.Error(err))
		}
		if err := t.WriteResponseFrame(op.Header(), resp); err != nil {
			s.stats.AddError(err)
			s.logger.Error("Failed to write response", zap.Error(err))
		}
	}
}
