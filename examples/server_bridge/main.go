package main

import (
	"os"
	"os/signal"

	"github.com/goburrow/serial"
	"github.com/rinzlerlabs/gomodbus/server"
	"github.com/rinzlerlabs/gomodbus/server/network/tcp"
	rtu "github.com/rinzlerlabs/gomodbus/server/serial/rtu"
	"go.uber.org/zap"
)

func main() {
	// We're using the "Production" logger here, which means INFO messages and above.
	// The RTU server can be very chatty because partial messages are very common.
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	handler := server.NewDefaultHandler(logger, 65535, 65535, 65535, 65535) // Create a shared handler for the 2 servers we are going to run

	// Open the serial port for the RTU server
	port, err := serial.Open(&serial.Config{
		Address:  "/dev/ttyUSB0",
		BaudRate: 19200,
		DataBits: 8,
		Parity:   "N",
		StopBits: 2,
	})
	if err != nil {
		logger.Error("Failed to open serial port", zap.Error(err))
		return
	}

	// Create the RTU server
	one, err := rtu.NewModbusServerWithHandler(logger, port, 91, handler)
	if err != nil {
		logger.Error("Failed to create RTU server", zap.Error(err))
		return
	}
	err = one.Start()
	if err != nil {
		logger.Error("Failed to start TCP server", zap.Error(err))
		return
	}
	defer one.Close() // Make sure we close it when we're done

	// Create the TCP server
	two, err := tcp.NewModbusServerWithHandler(logger, ":8502", handler)
	if err != nil {
		logger.Error("Failed to create TCP server", zap.Error(err))
		return
	}
	err = two.Start() // Start the server
	if err != nil {
		logger.Error("Failed to start TCP server", zap.Error(err))
		return
	}
	defer two.Close() // Don't forget to close it when we're done

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt) // Wait until Ctrl+C is pressed
	<-c
}
