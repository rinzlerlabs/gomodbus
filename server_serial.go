package gomodbus

import (
	"bufio"
	"io"
)

type modbusSerialServer struct {
	modbusServer
	address uint16
	port    io.ReadWriteCloser
	reader  *bufio.Reader
}
