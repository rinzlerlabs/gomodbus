package common

import "errors"

var (
	ErrInvalidPacket             = errors.New("invalid packet")
	ErrInvalidChecksum           = errors.New("invalid checksum")
	ErrWrittenLengthDoesNotMatch = errors.New("written length does not match")
	ErrUnknownFunctionCode       = errors.New("unknown function code")
	ErrShortWrite                = errors.New("short write")
	ErrTimeout                   = errors.New("timeout")
	ErrIgnorePacket              = errors.New("ignore packet")
	ErrNotOurAddress             = errors.New("not our address")
	ErrUnsupportedFunctionCode   = errors.New("unsupported function code")
	ErrInvalidFunctionCode       = errors.New("invalid function code")
	ErrInvalidData               = errors.New("invalid data")
	ErrInvalidAddress            = errors.New("invalid address")
	ErrInvalidCount              = errors.New("invalid count")
	ErrInvalidValue              = errors.New("invalid value")
	ErrResponseValueMismatch     = errors.New("response value mismatch")
	ErrResponseOffsetMismatch    = errors.New("response offset mismatch")
)
