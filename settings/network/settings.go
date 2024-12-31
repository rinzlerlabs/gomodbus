package network

import (
	"net/url"
	"time"

	"github.com/rinzlerlabs/gomodbus/common"
)

type NetworkSettings struct {
	Endpoint  *url.URL
	KeepAlive time.Duration
}

func (n *NetworkSettings) parseValuesFromURI(u *url.URL) error {
	if err := validateScheme(u); err != nil {
		return common.ErrInvalidScheme
	}
	n.Endpoint = u
	if err := parseFieldDurationFromURL(u, "keepAlive", &n.KeepAlive, 30*time.Second); err != nil {
		return err
	}
	return nil
}

type ClientSettings struct {
	NetworkSettings
	ResponseTimeout time.Duration
	DialTimeout     time.Duration
}

func (c *ClientSettings) parseValuesFromURI(u *url.URL) error {
	if err := c.NetworkSettings.parseValuesFromURI(u); err != nil {
		return err
	}
	if err := parseFieldDurationFromURL(u, "responseTimeout", &c.ResponseTimeout, 1*time.Second); err != nil {
		return err
	}
	if err := parseFieldDurationFromURL(u, "dialTimeout", &c.DialTimeout, 5*time.Second); err != nil {
		return err
	}
	return nil
}

type ServerSettings struct {
	NetworkSettings
}

func (s *ServerSettings) parseValuesFromURI(u *url.URL) error {
	if err := s.NetworkSettings.parseValuesFromURI(u); err != nil {
		return err
	}
	return nil
}

func parseFieldDurationFromURL(u *url.URL, field string, settingsField *time.Duration, defaultValue time.Duration) error {
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

func NewClientSettingsFromURI(uri string) (*ClientSettings, error) {
	if uri == "" {
		return nil, common.ErrURIIsNil
	}
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	if err := validateScheme(u); err != nil {
		return nil, common.ErrInvalidScheme
	}

	clientSettings := &ClientSettings{}
	if err := clientSettings.parseValuesFromURI(u); err != nil {
		return nil, err
	}

	return clientSettings, nil
}

func NewServerSettingsFromURI(uri string) (*ServerSettings, error) {
	if uri == "" {
		return nil, common.ErrURIIsNil
	}
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	if err := validateScheme(u); err != nil {
		return nil, common.ErrInvalidScheme
	}

	serverSettings := &ServerSettings{}
	if err := serverSettings.parseValuesFromURI(u); err != nil {
		return nil, err
	}

	return serverSettings, nil
}

func DefaultClientSettings(endpoint string) (*ClientSettings, error) {
	u, e := url.Parse(endpoint)
	if e != nil {
		return nil, e
	}
	if err := validateScheme(u); err != nil {
		return nil, common.ErrInvalidScheme
	}

	return &ClientSettings{
		NetworkSettings: NetworkSettings{
			Endpoint:  u,
			KeepAlive: 30 * time.Second,
		},
		ResponseTimeout: 5 * time.Second,
		DialTimeout:     5 * time.Second,
	}, nil
}

func NewClientSettings(endpoint string, dialTimeout, responseTimeout, keepAlive time.Duration) (*ClientSettings, error) {
	u, e := url.Parse(endpoint)
	if e != nil {
		return nil, e
	}
	if err := validateScheme(u); err != nil {
		return nil, common.ErrInvalidScheme
	}

	return &ClientSettings{
		NetworkSettings: NetworkSettings{
			Endpoint:  u,
			KeepAlive: keepAlive,
		},
		ResponseTimeout: responseTimeout,
		DialTimeout:     dialTimeout,
	}, nil
}

func validateScheme(u *url.URL) error {
	if u.Scheme != "tcp" && u.Scheme != "udp" {
		return common.ErrInvalidScheme
	}
	return nil
}
