package network

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewModbusClient(t *testing.T) {
	listener, err := net.Listen("tcp", ":8502")
	assert.NoError(t, err)
	defer listener.Close()
	endpoint := "tcp://:8502"
	client, err := NewModbusClient(zap.NewNop(), endpoint, 0)
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNewModbusClientWithContext(t *testing.T) {
	listener, err := net.Listen("tcp", ":8502")
	assert.NoError(t, err)
	defer listener.Close()
	endpoint := "tcp://:8502"
	client, err := NewModbusClientWithContext(context.Background(), zap.NewNop(), endpoint, 0)
	assert.NoError(t, err)
	assert.NotNil(t, client)
}
