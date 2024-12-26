package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rinzlerlabs/gomodbus/common"
	"github.com/rinzlerlabs/gomodbus/data"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestHandlerSaveState(t *testing.T) {
	logger := zaptest.NewLogger(t)
	handler := NewDefaultHandler(logger, 10, 10, 10, 10)
	err := handler.Save("test_data")
	assert.NoError(t, err)
	createdFiles, err := filepath.Glob("test_data/*.dat")
	assert.NoError(t, err)
	for _, v := range createdFiles {
		fileInfo, err := os.Stat(v)
		assert.NoError(t, err)
		logger.Info("File Stats", zap.String("filename", v), zap.Int64("Size", fileInfo.Size()))
	}
}

func TestHandlerLoadState(t *testing.T) {
	logger := zaptest.NewLogger(t)
	createdFiles, err := filepath.Glob("test_data/*.dat")
	assert.NoError(t, err)
	for _, v := range createdFiles {
		fileInfo, err := os.Stat(v)
		assert.NoError(t, err)
		logger.Info("File Stats", zap.String("filename", v), zap.Int64("Size", fileInfo.Size()))
	}
	tests := []struct {
		name                         string
		coilSize                     uint16
		expectedCoilsSize            int
		discreteInputSize            uint16
		expectedDiscreteInputsSize   int
		holdingRegisterSize          uint16
		expectedHoldingRegistersSize int
		inputRegisterSize            uint16
		expectedInputRegistersSize   int
	}{
		{"ExactRegisterSize", 10, 10, 10, 10, 10, 10, 10, 10},
		{"SmallerRegisterSize", 5, 10, 5, 10, 5, 10, 5, 10},
		{"LargerRegisterSize", 15, 15, 15, 15, 15, 15, 15, 15},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDefaultHandler(logger, tt.coilSize, tt.discreteInputSize, tt.holdingRegisterSize, tt.inputRegisterSize)
			err := handler.Load("test_data")
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCoilsSize, len(handler.(*DefaultHandler).Coils))
			assert.Equal(t, tt.expectedDiscreteInputsSize, len(handler.(*DefaultHandler).DiscreteInputs))
			assert.Equal(t, tt.expectedHoldingRegistersSize, len(handler.(*DefaultHandler).HoldingRegisters))
			assert.Equal(t, tt.expectedInputRegistersSize, len(handler.(*DefaultHandler).InputRegisters))
		})
	}
}

func TestHandlerReadCoils(t *testing.T) {
	logger := zaptest.NewLogger(t)
	tests := []struct {
		name          string
		coilCount     uint16
		readOffset    uint16
		readQuantity  uint16
		expectedError error
	}{
		{"Valid", 10, 0, 10, nil},
		{"InvalidStartOffset", 10, 10, 1, common.ErrIllegalDataAddress},
		{"InvalidQuantity", 10, 0, 11, common.ErrIllegalDataAddress},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDefaultHandler(logger, tt.coilCount, 10, 10, 10)
			req := data.NewReadCoilsRequest(tt.readOffset, tt.readQuantity)
			_, err := handler.ReadCoils(req)
			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
