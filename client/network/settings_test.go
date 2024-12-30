package network

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewClientSettings(t *testing.T) {
	tests := []struct {
		name            string
		uri             string
		responseTimeout time.Duration
		dialTimeout     time.Duration
		keepAlive       time.Duration
		scheme          string
		host            string
		err             error
	}{
		{
			name: "nil uri",
			uri:  "",
			err:  ErrURIIsNil,
		},
		{
			name: "wrong transport",
			uri:  "ascii://:502",
			err:  ErrInvalidScheme,
		},
		{
			name:            "Default values",
			uri:             "tcp://:502",
			err:             nil,
			responseTimeout: 1 * time.Second,
			dialTimeout:     5 * time.Second,
			keepAlive:       30 * time.Second,
			scheme:          "tcp",
			host:            ":502",
		},
		{
			name:            "Custom values",
			uri:             "tcp://:502?responseTimeout=2s&dialTimeout=10s&keepAlive=60s",
			err:             nil,
			responseTimeout: 2 * time.Second,
			dialTimeout:     10 * time.Second,
			keepAlive:       60 * time.Second,
			scheme:          "tcp",
			host:            ":502",
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
				if tt.responseTimeout != 0 {
					assert.Equal(t, tt.responseTimeout, settings.ResponseTimeout)
				}
				if tt.dialTimeout != 0 {
					assert.Equal(t, tt.dialTimeout, settings.DialTimeout)
				}
				if tt.keepAlive != 0 {
					assert.Equal(t, tt.keepAlive, settings.KeepAlive)
				}
				if tt.scheme != "" {
					assert.Equal(t, tt.scheme, settings.Endpoint.Scheme)
				}
				if tt.host != "" {
					assert.Equal(t, tt.host, settings.Endpoint.Host)
				}
			}
		})
	}
}
