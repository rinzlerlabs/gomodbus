package network

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewModbusClientFromSettings(t *testing.T) {
	listener, err := net.Listen("tcp", ":8502")
	assert.NoError(t, err)
	defer listener.Close()
	settings, err := DefaultClientSettings("tcp://:8502")
	assert.NoError(t, err)
	client, err := NewModbusClientFromSettings(zap.NewNop(), settings)
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNewModbusClientFromSettingsWithContext(t *testing.T) {
	listener, err := net.Listen("tcp", ":8502")
	assert.NoError(t, err)
	defer listener.Close()
	settings, err := DefaultClientSettings("tcp://:8502")
	assert.NoError(t, err)
	client, err := NewModbusClientFromSettingsWithContext(context.Background(), zap.NewNop(), settings)
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNewModbusClient(t *testing.T) {
	listener, err := net.Listen("tcp", ":8502")
	assert.NoError(t, err)
	defer listener.Close()

	uri := "tcp://:8502"
	client, err := NewModbusClient(zap.NewNop(), uri)
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNewModbusClientWithContext(t *testing.T) {
	ctx := context.Background()
	listener, err := net.Listen("tcp", ":8502")
	assert.NoError(t, err)
	defer listener.Close()

	uri := "tcp://:8502"
	client, err := NewModbusClientWithContext(ctx, zap.NewNop(), uri)
	assert.NoError(t, err)
	assert.NotNil(t, client)
}
