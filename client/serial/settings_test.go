package serial

import (
	"net/url"
	"testing"

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
		scheme   serialTransport
		device   string
		err      error
	}{
		{
			name: "nil uri",
			uri:  "",
			err:  ErrURIIsNil,
		},
		{
			name: "wrong transport",
			uri:  "tcp://:502",
			err:  ErrInvalidScheme,
		},
		{
			name:   "missing baud rate",
			uri:    "rtu:///dev/ttyUSB0",
			err:    ErrInvalidBaudRate,
			scheme: RTU,
		},
		{
			name:   "invalid baud rate",
			uri:    "rtu:///dev/ttyUSB0?baud=9600x",
			err:    ErrInvalidBaudRate,
			scheme: RTU,
		},
		{
			name:   "missing data bits",
			uri:    "rtu:///dev/ttyUSB0?baud=9600",
			err:    ErrInvalidDataBits,
			scheme: RTU,
		},
		{
			name:   "invalid data bits",
			uri:    "rtu:///dev/ttyUSB0?baud=9600&dataBits=8s",
			err:    ErrInvalidDataBits,
			scheme: RTU,
		},
		{
			name:   "missing parity",
			uri:    "rtu:///dev/ttyUSB0?baud=9600&dataBits=8",
			err:    ErrInvalidParity,
			scheme: RTU,
		},
		{
			name:   "invalid parity",
			uri:    "rtu:///dev/ttyUSB0?baud=9600&dataBits=8&parity=Z",
			err:    ErrInvalidParity,
			scheme: RTU,
		},
		{
			name:   "missing stop bits",
			uri:    "rtu:///dev/ttyUSB0?baud=9600&dataBits=8&parity=E",
			err:    ErrInvalidStopBits,
			scheme: RTU,
		},
		{
			name:   "invalid stop bits",
			uri:    "rtu:///dev/ttyUSB0?baud=9600&dataBits=8&parity=E&stopBits=2s",
			err:    ErrInvalidStopBits,
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
			var u *url.URL
			var err error
			if tt.uri == "" {
				u = nil
			} else {
				u, err = url.Parse(tt.uri)
			}
			assert.NoError(t, err)
			settings, err := NewClientSettingsFromURI(u)
			if err != nil {
				assert.ErrorIs(t, err, tt.err)
				assert.Nil(t, settings)
				return
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, settings)
				if tt.device != "" {
					assert.Equal(t, tt.device, settings.serialSettings.Device)
				}
				if tt.baud != 0 {
					assert.Equal(t, tt.baud, settings.serialSettings.Baud)
				}
				if tt.dataBits != 0 {
					assert.Equal(t, tt.dataBits, settings.serialSettings.DataBits)
				}
				if tt.parity != "" {
					assert.Equal(t, tt.parity, settings.serialSettings.Parity)
				}
				if tt.stopBits != 0 {
					assert.Equal(t, tt.stopBits, settings.serialSettings.StopBits)
				}
				if tt.scheme != "" {
					assert.Equal(t, tt.scheme, settings.serialSettings.Transport)
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
		scheme   serialTransport
		device   string
		err      error
	}{
		{
			name: "nil uri",
			uri:  "",
			err:  ErrURIIsNil,
		},
		{
			name: "wrong transport",
			uri:  "tcp://:502",
			err:  ErrInvalidScheme,
		},
		{
			name:   "missing baud rate",
			uri:    "rtu:///dev/ttyUSB0",
			err:    ErrInvalidBaudRate,
			scheme: RTU,
		},
		{
			name:   "invalid baud rate",
			uri:    "rtu:///dev/ttyUSB0?baud=9600x",
			err:    ErrInvalidBaudRate,
			scheme: RTU,
		},
		{
			name:   "missing data bits",
			uri:    "rtu:///dev/ttyUSB0?baud=9600",
			err:    ErrInvalidDataBits,
			scheme: RTU,
		},
		{
			name:   "invalid data bits",
			uri:    "rtu:///dev/ttyUSB0?baud=9600&dataBits=8s",
			err:    ErrInvalidDataBits,
			scheme: RTU,
		},
		{
			name:   "missing parity",
			uri:    "rtu:///dev/ttyUSB0?baud=9600&dataBits=8",
			err:    ErrInvalidParity,
			scheme: RTU,
		},
		{
			name:   "invalid parity",
			uri:    "rtu:///dev/ttyUSB0?baud=9600&dataBits=8&parity=Z",
			err:    ErrInvalidParity,
			scheme: RTU,
		},
		{
			name:   "missing stop bits",
			uri:    "rtu:///dev/ttyUSB0?baud=9600&dataBits=8&parity=E",
			err:    ErrInvalidStopBits,
			scheme: RTU,
		},
		{
			name:   "invalid stop bits",
			uri:    "rtu:///dev/ttyUSB0?baud=9600&dataBits=8&parity=E&stopBits=2s",
			err:    ErrInvalidStopBits,
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
			var u *url.URL
			var err error
			if tt.uri == "" {
				u = nil
			} else {
				u, err = url.Parse(tt.uri)
			}
			assert.NoError(t, err)
			settings, err := parseSerialSettingsFromUrl(u)
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
