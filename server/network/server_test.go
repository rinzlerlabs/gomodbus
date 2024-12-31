package network

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rinzlerlabs/gomodbus/server"
	settings "github.com/rinzlerlabs/gomodbus/settings/network"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func newModbusServerWithHandler(logger *zap.Logger, listener net.Listener, handler server.RequestHandler) (server.ModbusServer, error) {
	if handler == nil {
		return nil, errors.New("handler is required")
	}
	ctx, cancel := context.WithCancel(context.Background())
	url, err := url.Parse("tcp://localhost:502")
	if err != nil {
		return nil, err
	}
	return &modbusServer{
		logger:    logger,
		handler:   handler,
		cancelCtx: ctx,
		cancel:    cancel,
		listener:  listener,
		settings: &settings.ServerSettings{
			NetworkSettings: settings.NetworkSettings{
				Endpoint:  url,
				KeepAlive: 30 * time.Second,
			},
		},
		stats: server.NewServerStats(),
	}, nil
}

type timeoutError struct{}

func (t *timeoutError) Error() string {
	return "timeout"
}

func (t *timeoutError) Timeout() bool {
	return true
}

type testListener struct {
	mu        sync.Mutex
	readData  [][]byte
	writeData []byte
}

func (t *testListener) Accept() (net.Conn, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.readData) == 0 {
		return nil, &net.OpError{Err: &timeoutError{}}
	}
	r := t.readData[0]
	t.readData = t.readData[1:]

	return &testConnection{readData: r, listener: t}, nil
}

func (t *testListener) Close() error {
	return nil
}

func (t *testListener) Addr() net.Addr {
	return nil
}

type testConnection struct {
	readData []byte
	listener *testListener
}

// LocalAddr implements net.Conn.
func (t *testConnection) LocalAddr() net.Addr {
	panic("unimplemented")
}

// SetDeadline implements net.Conn.
func (*testConnection) SetDeadline(t time.Time) error {
	panic("unimplemented")
}

// SetReadDeadline implements net.Conn.
func (*testConnection) SetReadDeadline(t time.Time) error {
	panic("unimplemented")
}

// SetWriteDeadline implements net.Conn.
func (*testConnection) SetWriteDeadline(t time.Time) error {
	panic("unimplemented")
}

func (c *testConnection) Read(b []byte) (n int, err error) {
	if len(c.readData) == 0 {
		return 0, io.EOF
	}
	d, e := hex.DecodeString(string(c.readData))
	if e != nil {
		return 0, e
	}
	lenRead := copy(b, d)
	c.readData = c.readData[lenRead*2:]
	return lenRead, nil
}

func (c *testConnection) Write(b []byte) (n int, err error) {
	fmt.Printf("Write: %v\n", len(b))
	c.listener.writeData = b
	return len(b), nil
}

func (c *testConnection) Close() error {
	return nil
}

func (c *testConnection) RemoteAddr() net.Addr {
	return &addr{}
}

type addr struct{}

func (a *addr) Network() string {
	return "tcp"
}

func (a *addr) String() string {
	return "localhost:502"
}

func waitForWrite(listener *testListener, desiredLength int) {
	timeout := time.After(1 * time.Second)
	tick := time.Tick(10 * time.Millisecond)

	for {
		select {
		case <-timeout:
			return
		case <-tick:
			if len(listener.writeData)*2 == desiredLength {
				return
			}
		}
	}
}

func TestNilHandlerReturnsError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	listener := &testListener{}
	_, err := newModbusServerWithHandler(logger, listener, nil)
	assert.Error(t, err)
}

func TestAcceptRequest(t *testing.T) {
	logger := zaptest.NewLogger(t)
	listener := &testListener{
		readData: [][]byte{[]byte("0002000000060101000A000D")},
	}
	s, err := newModbusServerWithHandler(logger, listener, server.NewDefaultHandler(logger, 1024, 1024, 1024, 1024))
	assert.NoError(t, err)

	s.Start()
	assert.NoError(t, err)

	waitForWrite(listener, 11)

	err = s.Close()
	assert.NoError(t, err)

	adu := listener.writeData

	assert.Equal(t, "0002000000050101020000", strings.ToUpper(hex.EncodeToString(adu)))
}

func TestReadCoils(t *testing.T) {
	tests := []struct {
		name     string
		request  string
		coils    []bool
		response string
	}{
		{
			name:     "Valid",
			request:  "0002000000060101000A000D",
			response: "0002000000050101020A11",
			coils:    []bool{true, false, false, false, false, false, false, false, false, false, false, true, false, true, false, false, false, false, true, false, false, false, true, true, true, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			listener := &testListener{
				readData: [][]byte{[]byte(tt.request)},
			}
			handler := server.NewDefaultHandler(logger, 1024, 1024, 1024, 1024)
			if tt.coils != nil {
				handler.(*server.DefaultHandler).Coils = tt.coils
			}
			s, err := newModbusServerWithHandler(logger, listener, handler)
			assert.NoError(t, err)

			s.Start()
			assert.NoError(t, err)

			waitForWrite(listener, len(tt.response))

			err = s.Close()
			assert.NoError(t, err)
			assert.Equal(t, tt.response, strings.ToUpper(hex.EncodeToString(listener.writeData)))
		})
	}
}

func TestReadDiscreteInputs(t *testing.T) {
	tests := []struct {
		name     string
		request  string
		inputs   []bool
		response string
	}{
		{
			name:     "Valid",
			request:  "0002000000060102000A000D",
			inputs:   []bool{true, false, false, false, false, false, false, false, false, false, false, true, false, true, false, false, false, false, true, false, false, false, true, true, true, true},
			response: "0002000000050102020A11",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			listener := &testListener{
				readData: [][]byte{[]byte(tt.request)},
			}
			handler := server.NewDefaultHandler(logger, 1024, 1024, 1024, 1024)
			if tt.inputs != nil {
				handler.(*server.DefaultHandler).DiscreteInputs = tt.inputs
			}
			s, err := newModbusServerWithHandler(logger, listener, handler)
			assert.NoError(t, err)

			s.Start()
			assert.NoError(t, err)

			waitForWrite(listener, len(tt.response))

			err = s.Close()
			assert.NoError(t, err)
			assert.Equal(t, tt.response, strings.ToUpper(hex.EncodeToString(listener.writeData)))
		})
	}
}

func TestReadHoldingRegisters(t *testing.T) {
	tests := []struct {
		name      string
		request   string
		registers []uint16
		response  string
	}{
		{
			name:      "Valid",
			request:   "000200000006010300000002",
			registers: []uint16{0x0006, 0x0005, 0x0004, 0x0003, 0x0002, 0x0001, 0x0000},
			response:  "00020000000701030400060005",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			listener := &testListener{
				readData: [][]byte{[]byte(tt.request)},
			}
			handler := server.NewDefaultHandler(logger, 1024, 1024, 1024, 1024)
			if tt.registers != nil {
				handler.(*server.DefaultHandler).HoldingRegisters = tt.registers
			}
			s, err := newModbusServerWithHandler(logger, listener, handler)
			assert.NoError(t, err)

			s.Start()
			assert.NoError(t, err)

			waitForWrite(listener, len(tt.response))

			err = s.Close()
			assert.NoError(t, err)
			assert.Equal(t, tt.response, strings.ToUpper(hex.EncodeToString(listener.writeData)))
		})
	}
}

func TestReadInputRegisters(t *testing.T) {
	tests := []struct {
		name      string
		request   string
		registers []uint16
		response  string
	}{
		{
			name:      "Valid",
			request:   "000200000006010400000002",
			response:  "00020000000701040400060005",
			registers: []uint16{0x0006, 0x0005, 0x0004, 0x0003, 0x0002, 0x0001, 0x0000},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			listener := &testListener{
				readData: [][]byte{[]byte(tt.request)},
			}
			handler := server.NewDefaultHandler(logger, 1024, 1024, 1024, 1024)
			if tt.registers != nil {
				handler.(*server.DefaultHandler).InputRegisters = tt.registers
			}
			s, err := newModbusServerWithHandler(logger, listener, handler)
			assert.NoError(t, err)

			s.Start()
			assert.NoError(t, err)

			waitForWrite(listener, len(tt.response))

			err = s.Close()
			assert.NoError(t, err)
			assert.Equal(t, tt.response, strings.ToUpper(hex.EncodeToString(listener.writeData)))
		})
	}
}

func TestWriteSingleCoil(t *testing.T) {
	tests := []struct {
		name      string
		request   string
		response  string
		coilIndex uint16
		coilValue bool
	}{
		{
			name:      "Valid",
			request:   "0002000000060105000AFF00",
			response:  "0002000000060105000AFF00",
			coilIndex: 10,
			coilValue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			listener := &testListener{
				readData: [][]byte{[]byte(tt.request)},
			}
			handler := server.NewDefaultHandler(logger, 1024, 1024, 1024, 1024)
			s, err := newModbusServerWithHandler(logger, listener, handler)
			assert.NoError(t, err)

			s.Start()
			assert.NoError(t, err)

			waitForWrite(listener, len(tt.response))

			err = s.Close()
			assert.NoError(t, err)
			if tt.coilIndex > 0 {
				assert.Equal(t, tt.coilValue, handler.(*server.DefaultHandler).Coils[tt.coilIndex+1])
			}
			assert.Equal(t, tt.response, strings.ToUpper(hex.EncodeToString(listener.writeData)))
		})
	}
}

func TestWriteSingleRegister(t *testing.T) {
	tests := []struct {
		name          string
		request       string
		response      string
		registerIndex uint16
		registerValue uint16
	}{
		{
			name:          "Valid",
			request:       "000200000006010600100003",
			response:      "000200000006010600100003",
			registerIndex: 0x10,
			registerValue: 0x0003,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			listener := &testListener{
				readData: [][]byte{[]byte(tt.request)},
			}
			handler := server.NewDefaultHandler(logger, 1024, 1024, 1024, 1024)
			s, err := newModbusServerWithHandler(logger, listener, handler)
			assert.NoError(t, err)

			s.Start()
			assert.NoError(t, err)

			waitForWrite(listener, len(tt.response))

			err = s.Close()
			assert.NoError(t, err)
			if tt.registerIndex > 0 {
				assert.Equal(t, tt.registerValue, handler.(*server.DefaultHandler).HoldingRegisters[tt.registerIndex+1])
			}
			assert.Equal(t, tt.response, strings.ToUpper(hex.EncodeToString(listener.writeData)))
		})
	}
}

func TestWriteMultipleCoils(t *testing.T) {
	tests := []struct {
		name              string
		request           string
		response          string
		expectedRegisters []bool
	}{
		{
			name:              "Valid",
			request:           "00020000000A010F0000001803018307",
			response:          "000200000006010F00000018",
			expectedRegisters: []bool{true, false, false, false, false, false, false, false, true, true, false, false, false, false, false, true, true, true, true, false, false, false, false, false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			listener := &testListener{
				readData: [][]byte{[]byte(tt.request)},
			}
			handler := server.NewDefaultHandler(logger, 1024, 1024, 1024, 1024)
			s, err := newModbusServerWithHandler(logger, listener, handler)
			assert.NoError(t, err)

			s.Start()
			assert.NoError(t, err)

			waitForWrite(listener, len(tt.response))

			err = s.Close()
			assert.NoError(t, err)
			if tt.expectedRegisters != nil {
				assert.Equal(t, tt.expectedRegisters, handler.(*server.DefaultHandler).Coils[0:24])
			}
			assert.Equal(t, tt.response, strings.ToUpper(hex.EncodeToString(listener.writeData)))
		})
	}
}

func TestWriteMultipleRegisters(t *testing.T) {
	tests := []struct {
		name              string
		request           string
		response          string
		expectedRegisters []uint16
	}{
		{
			name:              "Valid",
			request:           "00020000000B0110000000020400040002",
			response:          "000200000006011000000002",
			expectedRegisters: []uint16{0x0004, 0x0002},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			listener := &testListener{
				readData: [][]byte{[]byte(tt.request)},
			}
			handler := server.NewDefaultHandler(logger, 1024, 1024, 1024, 1024)
			s, err := newModbusServerWithHandler(logger, listener, handler)
			assert.NoError(t, err)

			s.Start()
			assert.NoError(t, err)

			waitForWrite(listener, len(tt.response))

			err = s.Close()
			assert.NoError(t, err)
			if tt.expectedRegisters != nil {
				assert.Equal(t, tt.expectedRegisters, handler.(*server.DefaultHandler).HoldingRegisters[0:2])
			}
			assert.Equal(t, tt.response, strings.ToUpper(hex.EncodeToString(listener.writeData)))
		})
	}
}

func TestGenerateLrc(t *testing.T) {
	data, err := hex.DecodeString("0401020000")
	assert.NoError(t, err)
	var lrc byte
	// then the data
	for _, b := range data {
		lrc += byte(b)
	}
	// Two's complement
	lrc = ^lrc + 1
	t.Logf("%x", []byte{lrc})
}
