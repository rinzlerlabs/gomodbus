package main

import (
	"os"
	"os/signal"

	"github.com/rinzlerlabs/gomodbus/server"
	network "github.com/rinzlerlabs/gomodbus/server/network"
	rtu "github.com/rinzlerlabs/gomodbus/server/serial/rtu"
	network_settings "github.com/rinzlerlabs/gomodbus/settings/network"
	serial_settings "github.com/rinzlerlabs/gomodbus/settings/serial"
	"go.uber.org/zap"
)

func main() {
	// We're using the "Production" logger here, which means INFO messages and above.
	// The RTU server can be very chatty because partial messages are very common.
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	handler := server.NewDefaultHandler(logger, 65535, 65535, 65535, 65535) // Create a shared handler for the 2 servers we are going to run

	uri := "rtu:///dev/ttyUSB0?baud=19200&dataBits=8&parity=N&stopBits=2&address=91"
	rtu_settings, err := serial_settings.NewServerSettingsFromURI(uri)
	if err != nil {
		logger.Error("Failed to create server settings", zap.Error(err))
		return
	}
	// Create the RTU server
	one, err := rtu.NewModbusServerWithHandler(logger, rtu_settings, handler)
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
	uri = "tcp://:8502"
	tcp_settings, err := network_settings.NewServerSettingsFromURI(uri)
	if err != nil {
		logger.Error("Failed to create server settings", zap.Error(err))
		return
	}
	two, err := network.NewModbusServerWithHandler(logger, tcp_settings, handler)
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
