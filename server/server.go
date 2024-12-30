package server

type ModbusServer interface {
	Start() error
	Close() error
	IsRunning() bool
	Stats() *ServerStats
}
