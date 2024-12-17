package gomodbus

import "time"

type ModbusUDPServer struct {
	modbusServer
	Endpoints []string
	Timeout   time.Duration
}
