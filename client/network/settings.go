package network

import (
	"errors"
	"net/url"
	"time"
)

var (
	ErrURIIsNil      = errors.New("uri is nil")
	ErrInvalidScheme = errors.New("unsupported scheme")
)

type ClientSettings struct {
	Endpoint        *url.URL
	ResponseTimeout time.Duration
	DialTimeout     time.Duration
	KeepAlive       time.Duration
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

func NewClientSettingsFromURI(uri *url.URL) (*ClientSettings, error) {
	if uri == nil {
		return nil, ErrURIIsNil
	}
	if err := validateScheme(uri); err != nil {
		return nil, ErrInvalidScheme
	}
	clientSettings := &ClientSettings{
		Endpoint: uri,
	}

	if err := parseFieldDurationFromURL(uri, "responseTimeout", &clientSettings.ResponseTimeout, 1*time.Second); err != nil {
		return nil, err
	}
	if err := parseFieldDurationFromURL(uri, "dialTimeout", &clientSettings.DialTimeout, 5*time.Second); err != nil {
		return nil, err
	}
	if err := parseFieldDurationFromURL(uri, "keepAlive", &clientSettings.KeepAlive, 30*time.Second); err != nil {
		return nil, err
	}

	return clientSettings, nil
}

func DefaultClientSettings(endpoint string) (*ClientSettings, error) {
	u, e := url.Parse(endpoint)
	if e != nil {
		return nil, e
	}
	if err := validateScheme(u); err != nil {
		return nil, ErrInvalidScheme
	}

	return &ClientSettings{
		Endpoint:        u,
		ResponseTimeout: 5 * time.Second,
		DialTimeout:     5 * time.Second,
		KeepAlive:       30 * time.Second,
	}, nil
}

func NewClientSettings(endpoint string, dialTimeout, responseTimeout, keepAlive time.Duration) (*ClientSettings, error) {
	u, e := url.Parse(endpoint)
	if e != nil {
		return nil, e
	}
	if err := validateScheme(u); err != nil {
		return nil, ErrInvalidScheme
	}

	return &ClientSettings{
		Endpoint:        u,
		ResponseTimeout: responseTimeout,
		DialTimeout:     dialTimeout,
		KeepAlive:       keepAlive,
	}, nil
}

func validateScheme(u *url.URL) error {
	if u.Scheme != "tcp" && u.Scheme != "udp" {
		return ErrInvalidScheme
	}
	return nil
}
