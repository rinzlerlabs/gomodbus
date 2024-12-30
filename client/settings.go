package client

import "time"

type ClientSettings interface {
	Endpoint() string
	ResponseTimeout() time.Duration
}
