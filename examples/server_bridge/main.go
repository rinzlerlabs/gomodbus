package main

import (
	"os"
	"os/signal"

	"github.com/rinzlerlabs/gomodbus/server"
	"github.com/rinzlerlabs/gomodbus/server/network/tcp"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	handler := server.NewDefaultHandler(logger, 65535, 65535, 65535, 65535)
	one, err := tcp.NewModbusServerWithHandler(logger, ":8502", handler)
	if err != nil {
		logger.Error("Failed to create TCP server", zap.Error(err))
		return
	}
	err = one.Start()
	if err != nil {
		logger.Error("Failed to start TCP server", zap.Error(err))
		return
	}
	defer one.Stop()

	two, err := tcp.NewModbusServerWithHandler(logger, ":8503", handler)
	if err != nil {
		logger.Error("Failed to create TCP server", zap.Error(err))
		return
	}
	err = two.Start()
	if err != nil {
		logger.Error("Failed to start TCP server", zap.Error(err))
		return
	}
	defer two.Stop()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
