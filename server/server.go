package server

type ModbusServer interface {
	IsRunning() bool
	Start() error
	Stop() error
}
