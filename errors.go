package gomodbus

import "errors"

var (
	ErrInvalidPacket             = errors.New("invalid packet")
	ErrInvalidChecksum           = errors.New("invalid checksum")
	ErrWrittenLengthDoesNotMatch = errors.New("written length does not match")
	ErrNotOurAddress             = errors.New("not our address")
	ErrUnknownFunctionCode       = errors.New("unknown function code")
	ErrUnsupportedFunctionCode   = errors.New("unsupported function code")
	errIgnorePacket              = errors.New("ignore packet")
	errTimeout                   = errors.New("timeout")
)
