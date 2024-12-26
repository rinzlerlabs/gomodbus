package main

import (
	"time"

	"github.com/goburrow/serial"
	"github.com/rinzlerlabs/gomodbus/client/serial/ascii"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	settings := &serial.Config{
		Address:  "/dev/ttyUSB0",
		BaudRate: 19200,
		DataBits: 8,
		Parity:   "N",
		StopBits: 1,
	}

	modbusClient, err := ascii.NewModbusClient(logger, settings, 1*time.Second)
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
