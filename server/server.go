package server

import (
	"sync"

	"github.com/rinzlerlabs/gomodbus/data"
	"github.com/rinzlerlabs/gomodbus/transport"
)

type ModbusServer interface {
	Start() error
	Stop() error
	Close() error
	IsRunning() bool
	Stats() *ServerStats
}

func NewServerStats() *ServerStats {
	return &ServerStats{LastErrors: make([]error, 0)}
}

type ServerStats struct {
	TotalRequests                       uint64
	TotalErrors                         uint64
	TotalClients                        uint64
	TotalReadCoilsRequests              uint64
	TotalReadDiscreteInputsRequests     uint64
	TotalReadHoldingRegistersRequests   uint64
	TotalReadInputRegistersRequests     uint64
	TotalWriteSingleCoilRequests        uint64
	TotalWriteSingleRegisterRequests    uint64
	TotalWriteMultipleCoilsRequests     uint64
	TotalWriteMultipleRegistersRequests uint64
	LastErrors                          []error
	mu                                  sync.Mutex
}

func (s *ServerStats) AddRequest(txn transport.ModbusTransaction) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TotalRequests++
	switch txn.Frame().PDU().FunctionCode() {
	case data.ReadCoils:
		s.TotalReadCoilsRequests++
	case data.ReadDiscreteInputs:
		s.TotalReadDiscreteInputsRequests++
	case data.ReadHoldingRegisters:
		s.TotalReadHoldingRegistersRequests++
	case data.ReadInputRegisters:
		s.TotalReadInputRegistersRequests++
	case data.WriteSingleCoil:
		s.TotalWriteSingleCoilRequests++
	case data.WriteSingleRegister:
		s.TotalWriteSingleRegisterRequests++
	case data.WriteMultipleCoils:
		s.TotalWriteMultipleCoilsRequests++
	case data.WriteMultipleRegisters:
		s.TotalWriteMultipleRegistersRequests++
	}
}

func (s *ServerStats) AddError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TotalErrors++
	if len(s.LastErrors) > 10 {
		s.LastErrors = s.LastErrors[1:]
	}
	s.LastErrors = append(s.LastErrors, err)
}

func (s *ServerStats) AsMap() map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	return map[string]interface{}{
		"TotalRequests":                       s.TotalRequests,
		"TotalErrors":                         s.TotalErrors,
		"TotalClients":                        s.TotalClients,
		"TotalReadCoilsRequests":              s.TotalReadCoilsRequests,
		"TotalReadDiscreteInputsRequests":     s.TotalReadDiscreteInputsRequests,
		"TotalReadHoldingRegistersRequests":   s.TotalReadHoldingRegistersRequests,
		"TotalReadInputRegistersRequests":     s.TotalReadInputRegistersRequests,
		"TotalWriteSingleCoilRequests":        s.TotalWriteSingleCoilRequests,
		"TotalWriteSingleRegisterRequests":    s.TotalWriteSingleRegisterRequests,
		"TotalWriteMultipleCoilsRequests":     s.TotalWriteMultipleCoilsRequests,
		"TotalWriteMultipleRegistersRequests": s.TotalWriteMultipleRegistersRequests,
		"LastErrors":                          s.LastErrors,
	}
}
