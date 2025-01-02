package rtu

import (
	"io"
	"time"

	sp "github.com/goburrow/serial"
	"github.com/rinzlerlabs/gomodbus/common"
	"github.com/rinzlerlabs/gomodbus/server"
	"github.com/rinzlerlabs/gomodbus/server/serial"
	settings "github.com/rinzlerlabs/gomodbus/settings/serial"
	"github.com/rinzlerlabs/gomodbus/transport/serial/rtu"
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

	port, err := sp.Open(serverSettings.GetSerialPortConfig())
	if err != nil {
		logger.Error("Failed to open serial port", zap.Error(err))
		return nil, err
	}
	internalPort := newRTUSerialPort(port)
	transport := rtu.NewModbusServerTransport(internalPort, logger, serverSettings.Address)

	return serial.NewModbusSerialServerWithTransport(logger, serverSettings, handler, transport)
}

func newRTUSerialPort(port io.ReadWriteCloser) *rtuSerialPort {
	return &rtuSerialPort{
		port:         port,
		lastActivity: time.Now(),
	}
}

type rtuSerialPort struct {
	io.ReadWriteCloser
	port         io.ReadWriteCloser
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
