package udp

import (
	"context"
	"sync"
	"time"

	"github.com/rinzlerlabs/gomodbus/server"
	"go.uber.org/zap"
)

type modbusServer struct {
	handler   server.RequestHandler
	cancelCtx context.Context
	cancel    context.CancelFunc
	logger    *zap.Logger
	mu        sync.Mutex
	Endpoints []string
	Timeout   time.Duration
}
