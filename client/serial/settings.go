package serial

import (
	"errors"
	"net/url"
	"strconv"
	"time"

	sp "github.com/goburrow/serial"
)

var (
	ErrURIIsNil        = errors.New("uri is nil")
	ErrInvalidScheme   = errors.New("unsupported scheme")
	ErrInvalidBaudRate = errors.New("invalid baud rate")
	ErrInvalidDataBits = errors.New("invalid data bits")
	ErrInvalidParity   = errors.New("invalid parity")
	ErrInvalidStopBits = errors.New("invalid stop bits")
)

type serialTransport string

const (
	RTU   serialTransport = "rtu"
	ASCII serialTransport = "ascii"
)

type SerialSettings struct {
	Transport serialTransport
	Device    string
	Baud      int
	DataBits  int
	Parity    string
	StopBits  int
}

func (s *SerialSettings) ToPortConfig() *sp.Config {
	return &sp.Config{
		Address:  s.Device,
		BaudRate: s.Baud,
		DataBits: s.DataBits,
		Parity:   s.Parity,
		StopBits: s.StopBits,
	}
}

type ClientSettings struct {
	serialSettings  *SerialSettings
	responseTimeout time.Duration
}

func (c *ClientSettings) SerialSettings() *SerialSettings {
	return c.serialSettings
}

func (c *ClientSettings) ResponseTimeout() time.Duration {
	return c.responseTimeout
}

func (c *ClientSettings) Endpoint() string {
	return c.serialSettings.Device
}

func NewClientSettings(serialSettings *SerialSettings, responseTimeout time.Duration) *ClientSettings {
	return &ClientSettings{
		serialSettings:  serialSettings,
		responseTimeout: responseTimeout,
	}
}

func NewClientSettingsFromURI(uri *url.URL) (*ClientSettings, error) {
	if uri == nil {
		return nil, ErrURIIsNil
	}
	u, err := validateUrl(uri)
	if err != nil {
		return nil, err
	}
	clientSettings := &ClientSettings{}

	serialSettings, err := parseSerialSettingsFromUrl(u)
	if err != nil {
		return nil, err
	}
	clientSettings.serialSettings = serialSettings

	if responseTimeout := u.Query().Get("responseTimeout"); responseTimeout != "" {
		responseTimeout, err := time.ParseDuration(responseTimeout)
		if err != nil {
			return nil, err
		}
		clientSettings.responseTimeout = responseTimeout
	} else {
		clientSettings.responseTimeout = 1 * time.Second
	}

	return clientSettings, nil
}

func validateUrl(u *url.URL) (*url.URL, error) {
	if u.Scheme != "rtu" && u.Scheme != "ascii" {
		return nil, ErrInvalidScheme
	}
	return u, nil
}

func parseSerialSettingsFromUrl(u *url.URL) (*SerialSettings, error) {
	if u == nil {
		return nil, ErrURIIsNil
	}
	settings := &SerialSettings{
		Device: u.Path,
	}
	switch u.Scheme {
	case "rtu":
		settings.Transport = RTU
	case "ascii":
		settings.Transport = ASCII
	default:
		return nil, ErrInvalidScheme
	}
	if baud := u.Query().Get("baud"); baud != "" {
		baud, err := strconv.Atoi(baud)
		if err != nil {
			return nil, errors.Join(err, ErrInvalidBaudRate)
		}
		settings.Baud = baud
	} else {
		return nil, ErrInvalidBaudRate
	}

	if dataBits := u.Query().Get("dataBits"); dataBits != "" {
		dataBits, err := strconv.Atoi(dataBits)
		if err != nil {
			return nil, errors.Join(err, ErrInvalidDataBits)
		}
		settings.DataBits = dataBits
	} else {
		return nil, ErrInvalidDataBits
	}

	if parity := u.Query().Get("parity"); parity != "" {
		if parity != "N" && parity != "E" && parity != "O" {
			return nil, ErrInvalidParity
		}
		settings.Parity = parity
	} else {
		return nil, ErrInvalidParity
	}

	if stopBits := u.Query().Get("stopBits"); stopBits != "" {
		stopBits, err := strconv.Atoi(stopBits)
		if err != nil {
			return nil, errors.Join(err, ErrInvalidStopBits)
		}
		settings.StopBits = stopBits
	} else {
		return nil, ErrInvalidStopBits
	}

	return settings, nil
}
