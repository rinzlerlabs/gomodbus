package serial

import (
	"testing"

	"github.com/rinzlerlabs/gomodbus/common"
	"github.com/stretchr/testify/assert"
)

func TestNewClientSettings(t *testing.T) {
	tests := []struct {
		name     string
		uri      string
		baud     int
		dataBits int
		parity   string
		stopBits int
		scheme   SerialTransport
		device   string
		err      error
	}{
		{
			name: "nil uri",
			uri:  "",
			err:  common.ErrURIIsNil,
		},
		{
			name: "wrong transport",
			uri:  "tcp://:502",
			err:  common.ErrInvalidScheme,
		},
		{
			name:   "missing baud rate",
			uri:    "rtu:///dev/ttyUSB0",
			err:    common.ErrInvalidBaudRate,
			scheme: RTU,
		},
		{
			name:   "invalid baud rate",
			uri:    "rtu:///dev/ttyUSB0?baud=9600x",
			err:    common.ErrInvalidBaudRate,
			scheme: RTU,
		},
		{
			name:   "missing data bits",
			uri:    "rtu:///dev/ttyUSB0?baud=9600",
			err:    common.ErrInvalidDataBits,
			scheme: RTU,
		},
		{
			name:   "invalid data bits",
			uri:    "rtu:///dev/ttyUSB0?baud=9600&dataBits=8s",
			err:    common.ErrInvalidDataBits,
			scheme: RTU,
		},
		{
			name:   "missing parity",
			uri:    "rtu:///dev/ttyUSB0?baud=9600&dataBits=8",
			err:    common.ErrInvalidParity,
			scheme: RTU,
		},
		{
			name:   "invalid parity",
			uri:    "rtu:///dev/ttyUSB0?baud=9600&dataBits=8&parity=Z",
			err:    common.ErrInvalidParity,
			scheme: RTU,
		},
		{
			name:   "missing stop bits",
			uri:    "rtu:///dev/ttyUSB0?baud=9600&dataBits=8&parity=E",
			err:    common.ErrInvalidStopBits,
			scheme: RTU,
		},
		{
			name:   "invalid stop bits",
			uri:    "rtu:///dev/ttyUSB0?baud=9600&dataBits=8&parity=E&stopBits=2s",
			err:    common.ErrInvalidStopBits,
			scheme: RTU,
		},
		{
			name:     "valid uri",
			uri:      "rtu:///dev/ttyUSB1?baud=9600&dataBits=8&parity=E&stopBits=2",
			scheme:   RTU,
			device:   "/dev/ttyUSB1",
			baud:     9600,
			dataBits: 8,
			parity:   "E",
			stopBits: 2,
		},
		{
			name:     "valid uri",
			uri:      "ascii:///dev/ttyUSB1?baud=9600&dataBits=8&parity=E&stopBits=2",
			scheme:   ASCII,
			device:   "/dev/ttyUSB1",
			baud:     9600,
			dataBits: 8,
			parity:   "E",
			stopBits: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings, err := NewClientSettingsFromURI(tt.uri)
			if err != nil {
				assert.ErrorIs(t, err, tt.err)
				assert.Nil(t, settings)
				return
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, settings)
				if tt.device != "" {
					assert.Equal(t, tt.device, settings.Device)
				}
				if tt.baud != 0 {
					assert.Equal(t, tt.baud, settings.Baud)
				}
				if tt.dataBits != 0 {
					assert.Equal(t, tt.dataBits, settings.DataBits)
				}
				if tt.parity != "" {
					assert.Equal(t, tt.parity, settings.Parity)
				}
				if tt.stopBits != 0 {
					assert.Equal(t, tt.stopBits, settings.StopBits)
				}
				if tt.scheme != "" {
					assert.Equal(t, tt.scheme, settings.Transport)
				}
			}
		})
	}
}

func TestParseSerialSettingsFromUrl(t *testing.T) {
	tests := []struct {
		name     string
		uri      string
		baud     int
		dataBits int
		parity   string
		stopBits int
		scheme   SerialTransport
		device   string
		err      error
	}{
		{
			name: "nil uri",
			uri:  "",
			err:  common.ErrURIIsNil,
		},
		{
			name: "wrong transport",
			uri:  "tcp://:502",
			err:  common.ErrInvalidScheme,
		},
		{
			name:   "missing baud rate",
			uri:    "rtu:///dev/ttyUSB0",
			err:    common.ErrInvalidBaudRate,
			scheme: RTU,
		},
		{
			name:   "invalid baud rate",
			uri:    "rtu:///dev/ttyUSB0?baud=9600x",
			err:    common.ErrInvalidBaudRate,
			scheme: RTU,
		},
		{
			name:   "missing data bits",
			uri:    "rtu:///dev/ttyUSB0?baud=9600",
			err:    common.ErrInvalidDataBits,
			scheme: RTU,
		},
		{
			name:   "invalid data bits",
			uri:    "rtu:///dev/ttyUSB0?baud=9600&dataBits=8s",
			err:    common.ErrInvalidDataBits,
			scheme: RTU,
		},
		{
			name:   "missing parity",
			uri:    "rtu:///dev/ttyUSB0?baud=9600&dataBits=8",
			err:    common.ErrInvalidParity,
			scheme: RTU,
		},
		{
			name:   "invalid parity",
			uri:    "rtu:///dev/ttyUSB0?baud=9600&dataBits=8&parity=Z",
			err:    common.ErrInvalidParity,
			scheme: RTU,
		},
		{
			name:   "missing stop bits",
			uri:    "rtu:///dev/ttyUSB0?baud=9600&dataBits=8&parity=E",
			err:    common.ErrInvalidStopBits,
			scheme: RTU,
		},
		{
			name:   "invalid stop bits",
			uri:    "rtu:///dev/ttyUSB0?baud=9600&dataBits=8&parity=E&stopBits=2s",
			err:    common.ErrInvalidStopBits,
			scheme: RTU,
		},
		{
			name:     "valid uri",
			uri:      "rtu:///dev/ttyUSB1?baud=9600&dataBits=8&parity=E&stopBits=2",
			scheme:   RTU,
			device:   "/dev/ttyUSB1",
			baud:     9600,
			dataBits: 8,
			parity:   "E",
			stopBits: 2,
		},
		{
			name:     "valid uri",
			uri:      "ascii:///dev/ttyUSB1?baud=9600&dataBits=8&parity=E&stopBits=2",
			scheme:   ASCII,
			device:   "/dev/ttyUSB1",
			baud:     9600,
			dataBits: 8,
			parity:   "E",
			stopBits: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings, err := NewClientSettingsFromURI(tt.uri)
			if err != nil {
				assert.ErrorIs(t, err, tt.err)
				assert.Nil(t, settings)
				return
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, settings)
				if tt.device != "" {
					assert.Equal(t, tt.device, settings.Device)
				}
				if tt.baud != 0 {
					assert.Equal(t, tt.baud, settings.Baud)
				}
				if tt.dataBits != 0 {
					assert.Equal(t, tt.dataBits, settings.DataBits)
				}
				if tt.parity != "" {
					assert.Equal(t, tt.parity, settings.Parity)
				}
				if tt.stopBits != 0 {
					assert.Equal(t, tt.stopBits, settings.StopBits)
				}
				if tt.scheme != "" {
					assert.Equal(t, tt.scheme, settings.Transport)
				}
			}
		})
	}
}
