package ascii

import (
	"encoding/hex"
	"sync"
	"testing"
	"time"

	"github.com/rinzlerlabs/gomodbus/server"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

type testSerialPort struct {
	mu        sync.Mutex
	readData  []byte
	writeData []byte
}

func (t *testSerialPort) Read(b []byte) (n int, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.readData) == 0 {
		return 0, nil
	}
	lenRead := copy(b, t.readData)
	t.readData = t.readData[lenRead:]
	return lenRead, nil
}

func (t *testSerialPort) Write(b []byte) (n int, err error) {
	t.writeData = b
	return len(b), nil
}

func (t *testSerialPort) Close() error {
	return nil
}

func waitForWrite(port *testSerialPort, desiredLength int) {
	timeout := time.After(1 * time.Second)
	tick := time.Tick(10 * time.Millisecond)

	for {
		select {
		case <-timeout:
			return
		case <-tick:
			if len(port.writeData) == desiredLength {
				return
			}
		}
	}
}

func TestNilHandlerReturnsError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	port := &testSerialPort{}
	_, err := newModbusServerWithHandler(logger, port, 0x04, nil)
	assert.Error(t, err)
}

func TestAcceptRequest(t *testing.T) {
	logger := zaptest.NewLogger(t)
	port := &testSerialPort{
		readData: []byte(":0401000A000DE4\r\n"),
	}
	s, err := newModbusServerWithHandler(logger, port, 0x04, server.NewDefaultHandler(logger, 1024, 1024, 1024, 1024))
	assert.NoError(t, err)

	s.Start()
	assert.NoError(t, err)

	waitForWrite(port, 7)

	err = s.Stop()
	assert.NoError(t, err)

	adu := port.writeData

	assert.Equal(t, []byte(":0401020000F9\r\n"), adu)
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
			request:  ":0401000A000DE4\r\n",
			response: ":0401020A11DE\r\n",
			coils:    []bool{true, false, false, false, false, false, false, false, false, false, false, true, false, true, false, false, false, false, true, false, false, false, true, true, true, true},
		},
		{
			name:    "InvalidRequest_MissingTrailers",
			request: ":0401000A000DE4",
		},
		{
			name:    "InvalidRequest_InvalidStart",
			request: "0401000A000DE4\r\n",
		},
		{
			name:    "InvalidRequest_IvalidChecksum",
			request: ":0401000A000DE5\r\n",
		},
		{
			name:    "InvalidRequest_NotOurAddress",
			request: ":0501000A000DE3\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			handler := server.NewDefaultHandler(logger, 1024, 1024, 1024, 1024)
			if tt.coils != nil {
				handler.(*server.DefaultHandler).Coils = tt.coils
			}
			s, err := newModbusServerWithHandler(logger, port, 0x04, handler)
			assert.NoError(t, err)

			s.Start()
			assert.NoError(t, err)

			waitForWrite(port, len(tt.response))

			err = s.Stop()
			assert.NoError(t, err)
			assert.Equal(t, tt.response, string(port.writeData))
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
			request:  ":0402000A000DE3\r\n",
			inputs:   []bool{true, false, false, false, false, false, false, false, false, false, false, true, false, true, false, false, false, false, true, false, false, false, true, true, true, true},
			response: ":0402020A11DD\r\n",
		},
		{
			name:    "InvalidRequest_MissingTrailers",
			request: ":0402000A000DE3",
		},
		{
			name:    "InvalidRequest_InvalidStart",
			request: "0402000A000DE3\r\n",
		},
		{
			name:    "InvalidRequest_NotOurAddress",
			request: ":0502000A000DE2\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			handler := server.NewDefaultHandler(logger, 1024, 1024, 1024, 1024)
			if tt.inputs != nil {
				handler.(*server.DefaultHandler).DiscreteInputs = tt.inputs
			}
			s, err := newModbusServerWithHandler(logger, port, 0x04, handler)
			assert.NoError(t, err)

			s.Start()
			assert.NoError(t, err)

			waitForWrite(port, len(tt.response))

			err = s.Stop()
			assert.NoError(t, err)
			assert.Equal(t, tt.response, string(port.writeData))
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
			request:   ":040300000002F7\r\n",
			registers: []uint16{0x0006, 0x0005, 0x0004, 0x0003, 0x0002, 0x0001, 0x0000},
			response:  ":04030400060005EA\r\n",
		},
		{
			name:    "InvalidRequest_MissingTrailers",
			request: ":040300000002F7",
		},
		{
			name:    "InvalidRequest_InvalidStart",
			request: "040300000002F7\r\n",
		},
		{
			name:    "InvalidRequest_NotOurAddress",
			request: ":010300000002FA\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			handler := server.NewDefaultHandler(logger, 1024, 1024, 1024, 1024)
			if tt.registers != nil {
				handler.(*server.DefaultHandler).HoldingRegisters = tt.registers
			}
			s, err := newModbusServerWithHandler(logger, port, 0x04, handler)
			assert.NoError(t, err)

			s.Start()
			assert.NoError(t, err)

			waitForWrite(port, len(tt.response))

			err = s.Stop()
			assert.NoError(t, err)
			assert.Equal(t, tt.response, string(port.writeData))
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
			request:   ":040400000002F6\r\n",
			response:  ":04040400060005E9\r\n",
			registers: []uint16{0x0006, 0x0005, 0x0004, 0x0003, 0x0002, 0x0001, 0x0000},
		},
		{
			name:    "InvalidRequest_MissingTrailers",
			request: ":040400000002F6",
		},
		{
			name:    "InvalidRequest_InvalidStart",
			request: "040400000002F6\r\n",
		},
		{
			name:    "InvalidRequest_NotOurAddress",
			request: ":050400000002F5\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			handler := server.NewDefaultHandler(logger, 1024, 1024, 1024, 1024)
			if tt.registers != nil {
				handler.(*server.DefaultHandler).InputRegisters = tt.registers
			}
			s, err := newModbusServerWithHandler(logger, port, 0x04, handler)
			assert.NoError(t, err)

			s.Start()
			assert.NoError(t, err)

			waitForWrite(port, len(tt.response))

			err = s.Stop()
			assert.NoError(t, err)
			assert.Equal(t, tt.response, string(port.writeData))
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
			request:   ":0405000AFF00EE\r\n",
			response:  ":0405000AFF00EE\r\n",
			coilIndex: 10,
			coilValue: true,
		},
		{
			name:    "InvalidRequest_MissingTrailers",
			request: ":0405000AFF00EE",
		},
		{
			name:    "InvalidRequest_InvalidStart",
			request: "0405000AFF00EE\r\n",
		},
		{
			name:    "InvalidRequest_NotOurAddress",
			request: ":0505000AFF00ED\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			handler := server.NewDefaultHandler(logger, 1024, 1024, 1024, 1024)
			s, err := newModbusServerWithHandler(logger, port, 0x04, handler)
			assert.NoError(t, err)

			s.Start()
			assert.NoError(t, err)

			waitForWrite(port, len(tt.response))

			err = s.Stop()
			assert.NoError(t, err)
			if tt.coilIndex > 0 {
				assert.Equal(t, tt.coilValue, handler.(*server.DefaultHandler).Coils[tt.coilIndex+1])
			}
			assert.Equal(t, tt.response, string(port.writeData))
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
			request:       ":040600100003E3\r\n",
			response:      ":040600100003E3\r\n",
			registerIndex: 0x10,
			registerValue: 0x0003,
		},
		{
			name:    "InvalidRequest_MissingTrailers",
			request: ":040600100003E3",
		},
		{
			name:    "InvalidRequest_InvalidStart",
			request: "040600100003E3\r\n",
		},
		{
			name:    "InvalidRequest_NotOurAddress",
			request: ":050600100003E2\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			handler := server.NewDefaultHandler(logger, 1024, 1024, 1024, 1024)
			s, err := newModbusServerWithHandler(logger, port, 0x04, handler)
			assert.NoError(t, err)

			s.Start()
			assert.NoError(t, err)

			waitForWrite(port, len(tt.response))

			err = s.Stop()
			assert.NoError(t, err)
			if tt.registerIndex > 0 {
				assert.Equal(t, tt.registerValue, handler.(*server.DefaultHandler).HoldingRegisters[tt.registerIndex+1])
			}
			assert.Equal(t, tt.response, string(port.writeData))
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
			request:           ":040F000000180301830747\r\n",
			response:          ":040F00000018D5\r\n",
			expectedRegisters: []bool{true, false, false, false, false, false, false, false, true, true, false, false, false, false, false, true, true, true, true, false, false, false, false, false},
		},
		{
			name:    "InvalidRequest_MissingTrailers",
			request: ":040F000000180301830747",
		},
		{
			name:    "InvalidRequest_InvalidStart",
			request: "040F000000180301830747\r\n",
		},
		{
			name:    "InvalidRequest_NotOurAddress",
			request: ":050F000000180301830746\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			handler := server.NewDefaultHandler(logger, 1024, 1024, 1024, 1024)
			s, err := newModbusServerWithHandler(logger, port, 0x04, handler)
			assert.NoError(t, err)

			s.Start()
			assert.NoError(t, err)

			waitForWrite(port, len(tt.response))

			err = s.Stop()
			assert.NoError(t, err)
			if tt.expectedRegisters != nil {
				assert.Equal(t, tt.expectedRegisters, handler.(*server.DefaultHandler).Coils[0:24])
			}
			assert.Equal(t, tt.response, string(port.writeData))
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
			request:           ":0410000000020400040002E0\r\n",
			response:          ":041000000002EA\r\n",
			expectedRegisters: []uint16{0x0004, 0x0002},
		},
		{
			name:    "InvalidRequest_MissingTrailers",
			request: ":0410000000020400040002E0",
		},
		{
			name:    "InvalidRequest_InvalidStart",
			request: "0410000000020400040002E0\r\n",
		},
		{
			name:    "InvalidRequest_NotOurAddress",
			request: ":0510000000020400040002DF\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			logger := zaptest.NewLogger(t)
			port := &testSerialPort{
				readData: []byte(tt.request),
			}
			handler := server.NewDefaultHandler(logger, 1024, 1024, 1024, 1024)
			s, err := newModbusServerWithHandler(logger, port, 0x04, handler)
			assert.NoError(t, err)

			s.Start()
			assert.NoError(t, err)

			waitForWrite(port, len(tt.response))

			err = s.Stop()
			assert.NoError(t, err)
			if tt.expectedRegisters != nil {
				assert.Equal(t, tt.expectedRegisters, handler.(*server.DefaultHandler).HoldingRegisters[0:2])
			}
			assert.Equal(t, tt.response, string(port.writeData))
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
