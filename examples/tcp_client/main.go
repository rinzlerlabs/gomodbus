package main

import (
	"time"

	"github.com/rinzlerlabs/gomodbus/client"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	client, err := client.NewModbusTCPClient(logger, "100.103.61.25:502", 10*time.Second)
	if err != nil {
		logger.Error("Failed to create modbus client", zap.Error(err))
		return
	}
	defer client.Close()
	// registers, err := client.ReadHoldingRegisters(1, 0, 8)
	// if err != nil {
	// 	logger.Error("Failed to read holding registers", zap.Error(err))
	// 	return
	// }

	// logger.Info("Read holding registers", zap.Any("registers", registers))

	coils, err := client.ReadCoils(1, 1, 16)
	if err != nil {
		logger.Error("Failed to read coils", zap.Error(err))
		return
	}

	logger.Info("Read coils", zap.Any("coils", coils))

	// coils, err = client.ReadCoils(1, 1, 16)
	// if err != nil {
	// 	logger.Error("Failed to read coils", zap.Error(err))
	// 	return
	// }

	// logger.Info("Read coils", zap.Any("coils", coils))

	// coils, err = client.ReadCoils(1, 16, 2)
	// if err != nil {
	// 	logger.Error("Failed to read coils", zap.Error(err))
	// 	return
	// }

	logger.Info("Read coils", zap.Any("coils", coils))
}
