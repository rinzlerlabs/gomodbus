package main

import (
	"os"
	"os/signal"

	"github.com/goburrow/serial"
	"github.com/rinzlerlabs/gomodbus/server"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	rtuPort, err := serial.Open(&serial.Config{
		Address:  "/dev/ttyUSB0",
		BaudRate: 19200,
		DataBits: 8,
		Parity:   "N",
		StopBits: 2,
	})
	if err != nil {
		logger.Error("Failed to open port", zap.Error(err))
		return
	}
	asciiPort, err := serial.Open(&serial.Config{
		Address:  "/dev/ttyUSB1",
		BaudRate: 19200,
		DataBits: 7,
		Parity:   "N",
		StopBits: 2,
	})
	if err != nil {
		logger.Error("Failed to open port", zap.Error(err))
		return
	}

	handler := server.NewDefaultHandler(logger, 65535, 65535, 65535, 65535)
	rtu, err := server.NewModbusRTUServerWithHandler(logger, rtuPort, 91, handler)
	if err != nil {
		logger.Error("Failed to create RTU server", zap.Error(err))
		return
	}
	err = rtu.Start()
	if err != nil {
		logger.Error("Failed to start RTU server", zap.Error(err))
		return
	}
	defer rtu.Stop()

	ascii, err := server.NewModbusASCIIServerWithHandler(logger, asciiPort, 91, handler)
	if err != nil {
		logger.Error("Failed to create ASCII server", zap.Error(err))
		return
	}
	err = ascii.Start()
	if err != nil {
		logger.Error("Failed to start ASCII server", zap.Error(err))
		return
	}
	defer ascii.Stop()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
