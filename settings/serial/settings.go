package serial

import (
	"errors"
	"net/url"
	"strconv"
	"time"

	sp "github.com/goburrow/serial"
	"github.com/rinzlerlabs/gomodbus/common"
)

var (
	ErrInvalidBaudRate = errors.New("invalid baud rate")
	ErrInvalidDataBits = errors.New("invalid data bits")
	ErrInvalidParity   = errors.New("invalid parity")
	ErrInvalidStopBits = errors.New("invalid stop bits")
)

type SerialTransport string

const (
	RTU   SerialTransport = "rtu"
	ASCII SerialTransport = "ascii"
)

var (
	validParityValues = []string{"N", "E", "O"}
)

func (s *SerialSettings) GetSerialPortConfig() *sp.Config {
	return &sp.Config{
		Address:  s.Device,
		BaudRate: s.Baud,
		DataBits: s.DataBits,
		Parity:   s.Parity,
		StopBits: s.StopBits,
	}
}

type SerialSettings struct {
	Transport SerialTransport
	Device    string
	Baud      int
	DataBits  int
	Parity    string
	StopBits  int
}

func (s *SerialSettings) parseValuesFromURI(u *url.URL) error {
	if u == nil {
		return common.ErrURIIsNil
	}
	s.Device = u.Path
	switch u.Scheme {
	case "rtu":
		s.Transport = RTU
	case "ascii":
		s.Transport = ASCII
	default:
		return common.ErrInvalidScheme
	}
	if err := parseIntFieldFromURI(u, "baud", &s.Baud); err != nil {
		return errors.Join(err, ErrInvalidBaudRate)
	}

	if err := parseIntFieldFromURI(u, "dataBits", &s.DataBits); err != nil {
		return errors.Join(err, ErrInvalidDataBits)
	}

	if err := parseStringFieldFromURI(u, "parity", &s.Parity, validParityValues); err != nil {
		return errors.Join(err, ErrInvalidParity)
	}

	if err := parseIntFieldFromURI(u, "stopBits", &s.StopBits); err != nil {
		return errors.Join(err, ErrInvalidStopBits)
	}

	return nil
}

type ClientSettings struct {
	SerialSettings
	ResponseTimeout time.Duration
}

func (c *ClientSettings) parseValuesFromURI(u *url.URL) error {
	if err := c.SerialSettings.parseValuesFromURI(u); err != nil {
		return err
	}
	if err := parseFieldDurationFromURI(u, "responseTimeout", &c.ResponseTimeout, 1*time.Second); err != nil {
		return err
	}
	return nil
}

type ServerSettings struct {
	SerialSettings
	Address uint16
}

func (s *ServerSettings) parseValuesFromURI(u *url.URL) error {
	if err := s.SerialSettings.parseValuesFromURI(u); err != nil {
		return err
	}
	if err := parseUInt16FieldFromURI(u, "address", &s.Address); err != nil {
		return err
	}
	return nil
}

func parseFieldDurationFromURI(u *url.URL, field string, settingsField *time.Duration, defaultValue time.Duration) error {
	if value := u.Query().Get(field); value != "" {
		parsedValue, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*settingsField = parsedValue
	} else {
		*settingsField = defaultValue
	}
	return nil
}

func parseIntFieldFromURI(u *url.URL, field string, settingsField *int) error {
	if value := u.Query().Get(field); value != "" {
		value, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		*settingsField = value
		return nil
	}
	return common.ErrMissingValue
}

func parseUInt16FieldFromURI(u *url.URL, field string, settingsField *uint16) error {
	if value := u.Query().Get(field); value != "" {
		value, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		*settingsField = uint16(value)
		return nil
	}
	return common.ErrMissingValue
}

func parseStringFieldFromURI(u *url.URL, field string, settingsField *string, validValues []string) error {
	if value := u.Query().Get(field); value != "" {
		for _, v := range validValues {
			if v == value {
				*settingsField = value
				return nil
			}
		}
		return common.ErrInvalidValue
	}
	return common.ErrMissingValue
}

func NewClientSettingsFromURI(uri string) (*ClientSettings, error) {
	if uri == "" {
		return nil, common.ErrURIIsNil
	}
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	err = validateUrl(u)
	if err != nil {
		return nil, err
	}
	settings := &ClientSettings{}
	if err := settings.parseValuesFromURI(u); err != nil {
		return nil, err
	}

	return settings, nil
}

func NewServerSettingsFromURI(uri string) (*ServerSettings, error) {
	if uri == "" {
		return nil, common.ErrURIIsNil
	}
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	err = validateUrl(u)
	if err != nil {
		return nil, err
	}
	settings := &ServerSettings{}
	if err := settings.parseValuesFromURI(u); err != nil {
		return nil, err
	}

	return settings, nil
}

func validateUrl(u *url.URL) error {
	if u.Scheme != "rtu" && u.Scheme != "ascii" {
		return common.ErrInvalidScheme
	}
	return nil
}
