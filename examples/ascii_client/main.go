package main

import (
	"github.com/rinzlerlabs/gomodbus/client/serial/ascii"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	uri := "rtu:///dev/ttyUSB0?baud=19200&dataBits=7&parity=N&stopBits=1&responseTimeout=1s"

	modbusClient, err := ascii.NewModbusClient(logger, uri)
	if err != nil {
		logger.Error("Failed to create modbus client", zap.Error(err))
		return
	}
	defer modbusClient.Close()

	coils, err := modbusClient.ReadCoils(91, 0, 16)
	if err != nil {
		logger.Error("Failed to read coils", zap.Error(err))
		return
	}
	logger.Info("Read coils", zap.Any("coils", coils))
}
