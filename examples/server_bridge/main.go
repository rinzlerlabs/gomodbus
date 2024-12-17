package main

import (
	"os"
	"os/signal"

	"github.com/rinzlerlabs/gomodbus"
	"github.com/tarm/serial"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	rtuPort, err := serial.OpenPort(&serial.Config{
		Name:     "/dev/ttyUSB0",
		Baud:     19200,
		Size:     8,
		Parity:   serial.ParityNone,
		StopBits: serial.Stop1,
	})
	if err != nil {
		logger.Error("Failed to open port", zap.Error(err))
		return
	}
	asciiPort, err := serial.OpenPort(&serial.Config{
		Name:     "/dev/ttyUSB1",
		Baud:     19200,
		Size:     7,
		Parity:   serial.ParityNone,
		StopBits: serial.Stop1,
	})
	if err != nil {
		logger.Error("Failed to open port", zap.Error(err))
		return
	}

	handler := gomodbus.NewDefaultHandler(logger, 65535, 65535, 65535, 65535)
	rtu, err := gomodbus.NewModbusRTUServerWithHandler(logger, rtuPort, 91, handler)
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

	ascii, err := gomodbus.NewModbusASCIIServerWithHandler(logger, asciiPort, 91, handler)
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