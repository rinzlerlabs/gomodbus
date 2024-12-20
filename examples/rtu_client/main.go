package main

import (
	"time"

	"github.com/goburrow/serial"
	"github.com/rinzlerlabs/gomodbus/client"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	port, err := serial.Open(&serial.Config{
		Address:  "/dev/ttyUSB0",
		BaudRate: 19200,
		DataBits: 8,
		Parity:   "N",
		StopBits: 1,
	})
	if err != nil {
		logger.Error("Failed to open serial port", zap.Error(err))
		return
	}

	modbusClient := client.NewModbusRTUClient(logger, port, 1*time.Second)
	defer modbusClient.Close()

	coils, err := modbusClient.ReadCoils(91, 0, 16)
	if err != nil {
		logger.Error("Failed to read coils", zap.Error(err))
		return
	}
	logger.Info("Read coils", zap.Any("coils", coils))
}
