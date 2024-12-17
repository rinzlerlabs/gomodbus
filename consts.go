package gomodbus

type BaudRate int

const (
	BaudRate9600   BaudRate = 9600
	BaudRate19200  BaudRate = 19200
	BaudRate38400  BaudRate = 38400
	BaudRate57600  BaudRate = 57600
	BaudRate115200 BaudRate = 115200
)

type DataBits int

const (
	DataBits7 DataBits = 7
	DataBits8 DataBits = 8
)

type Parity string

const (
	ParityNone Parity = "N"
	ParityOdd  Parity = "O"
	ParityEven Parity = "E"
)

type StopBits int

const (
	StopBits1 StopBits = 1
	StopBits2 StopBits = 2
)
