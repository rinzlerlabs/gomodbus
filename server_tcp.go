package gomodbus

import "time"

type ModbusTCPServer struct {
	modbusServer
	Endpoints []string
	Timeout   time.Duration
}
