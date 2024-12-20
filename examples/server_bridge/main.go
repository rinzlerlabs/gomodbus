package main

import (
	"os"
	"os/signal"

	"github.com/rinzlerlabs/gomodbus/server"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	handler := server.NewDefaultHandler(logger, 65535, 65535, 65535, 65535)
	one, err := server.NewModbusTCPServerWithHandler(logger, ":8502", handler)
	if err != nil {
		logger.Error("Failed to create RTU server", zap.Error(err))
		return
	}
	err = one.Start()
	if err != nil {
		logger.Error("Failed to start RTU server", zap.Error(err))
		return
	}
	defer one.Stop()

	two, err := server.NewModbusTCPServerWithHandler(logger, ":8503", handler)
	if err != nil {
		logger.Error("Failed to create ASCII server", zap.Error(err))
		return
	}
	err = two.Start()
	if err != nil {
		logger.Error("Failed to start ASCII server", zap.Error(err))
		return
	}
	defer two.Stop()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
