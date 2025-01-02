package network

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"

	"github.com/rinzlerlabs/gomodbus/common"
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
		return nil, common.ErrHandlerRequired
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
		return common.ErrHandlerRequired
	}

	if s.listener == nil {
		listener, err := net.Listen(s.settings.Endpoint.Scheme, s.settings.Endpoint.Host)
		if err != nil {
			s.logger.Error("Failed to listen", zap.Error(err))
			return err
		}
		s.listener = listener
	}

	s.logger.Info("Starting Modbus TCP server", zap.String("endpoint", s.settings.Endpoint.Host), zap.String("port", s.settings.Endpoint.Port()))
	s.isRunning = true
	go s.run()
	return nil
}

func (s *modbusServer) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	s.logger.Info("Closing Modbus TCP server")
	if s.listener != nil {
		err = s.listener.Close()
		if err != nil {
			s.logger.Error("Error closing listener", zap.Error(err))
		}
	} else {
		s.logger.Info("Listener is nil, did the server fully start?")
	}
	if s.cancel != nil {
		defer func() { s.cancel = nil }()
		s.cancel()
	} else {
		s.logger.Info("Cancel function is nil, did the server fully start?")
	}
	s.logger.Info("Waiting for all clients to disconnect")
	s.wg.Wait()
	s.logger.Info("All clients disconnected")
	return err
}

func (s *modbusServer) Stats() *server.ServerStats {
	return s.stats
}

func (s *modbusServer) run() {
	s.logger.Info("Modbus TCP server started")
	for {
		select {
		case <-s.cancelCtx.Done():
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					s.logger.Debug("Timeout accepting connection, this is more or less expected")
					continue
				}
				if errors.Is(err, net.ErrClosed) {
					s.logger.Info("Listener closed")
					return
				}
				s.logger.Error("Failed to accept connection", zap.Error(err))
				continue
			}
			s.logger.Info("Client connected", zap.String("remote", conn.RemoteAddr().String()))
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
		if errors.Is(err, io.EOF) {
			s.logger.Info("Client disconnected, cleaning up transport and client", zap.String("remote", conn.RemoteAddr().String()), zap.Error(err))
			return
		} else if errors.Is(err, context.Canceled) {
			s.logger.Debug("Server context canceled, cleaning up transport and client", zap.String("remote", conn.RemoteAddr().String()), zap.Error(err))
			return
		} else if err != nil {
			s.logger.Error("Failed to accept request", zap.Error(err))
			continue
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
