package main

import (
	"github.com/rinzlerlabs/gomodbus/client/serial/rtu"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	url := "rtu:///dev/ttyUSB0?baud=19200&dataBits=8&parity=N&stopBits=1&responseTimeout=1s"
	modbusClient, err := rtu.NewModbusClient(logger, url)
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
