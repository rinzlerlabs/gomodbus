package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadCoilsResponse_Bytes(t *testing.T) {
	values := []bool{true, false, false, false, false, false, false, true, true, false, false, false, false, false, false, true}
	response := ReadCoilsResponse{Values: values}
	expected := []byte{0x02, 0x81, 0x81}
	result := response.Bytes()
	assert.Equal(t, expected, result)
}

func TestReadCoilsRequest_Bytes(t *testing.T) {
	request := ReadCoilsRequest{Offset: 0, Count: 16}
	expected := []byte{0x00, 0x00, 0x00, 0x10}
	result := request.Bytes()
	assert.Equal(t, expected, result)
}

func TestReadDiscreteInputsResponse_Bytes(t *testing.T) {
	values := []bool{true, false, false, false, false, false, false, true, true, false, false, false, false, false, false, true}
	response := ReadDiscreteInputsResponse{Values: values}
	expected := []byte{0x02, 0x81, 0x81}
	result := response.Bytes()
	assert.Equal(t, expected, result)
}

func TestReadDiscreteInputsRequest_Bytes(t *testing.T) {
	request := ReadDiscreteInputsRequest{Offset: 0, Count: 16}
	expected := []byte{0x00, 0x00, 0x00, 0x10}
	result := request.Bytes()
	assert.Equal(t, expected, result)
}

func TestReadHoldingRegistersResponse_Bytes(t *testing.T) {
	values := []uint16{0x0001, 0x0002, 0x0003, 0x0004}
	response := ReadHoldingRegistersResponse{Values: values}
	expected := []byte{0x08, 0x00, 0x01, 0x00, 0x02, 0x00, 0x03, 0x00, 0x04}
	result := response.Bytes()
	assert.Equal(t, expected, result)
}

func TestReadHoldingRegistersRequest_Bytes(t *testing.T) {
	request := ReadHoldingRegistersRequest{Offset: 0, Count: 4}
	expected := []byte{0x00, 0x00, 0x00, 0x04}
	result := request.Bytes()
	assert.Equal(t, expected, result)
}

func TestReadInputRegistersResponse_Bytes(t *testing.T) {
	values := []uint16{0x0001, 0x0002, 0x0003, 0x0004}
	response := ReadInputRegistersResponse{Values: values}
	expected := []byte{0x08, 0x00, 0x01, 0x00, 0x02, 0x00, 0x03, 0x00, 0x04}
	result := response.Bytes()
	assert.Equal(t, expected, result)
}

func TestReadInputRegistersRequest_Bytes(t *testing.T) {
	request := ReadInputRegistersRequest{Offset: 0, Count: 4}
	expected := []byte{0x00, 0x00, 0x00, 0x04}
	result := request.Bytes()
	assert.Equal(t, expected, result)
}

func TestWriteSingleCoilResponse_Bytes(t *testing.T) {
	response := WriteSingleCoilResponse{Offset: 0, Value: true}
	expected := []byte{0x00, 0x00, 0xFF, 0x00}
	result := response.Bytes()
	assert.Equal(t, expected, result)
}

func TestWriteSingleCoilRequest_Bytes(t *testing.T) {
	request := WriteSingleCoilRequest{Offset: 0, Value: true}
	expected := []byte{0x00, 0x00, 0xFF, 0x00}
	result := request.Bytes()
	assert.Equal(t, expected, result)
}

func TestWriteSingleRegisterResponse_Bytes(t *testing.T) {
	response := WriteSingleRegisterResponse{Offset: 0, Value: 0x0001}
	expected := []byte{0x00, 0x00, 0x00, 0x01}
	result := response.Bytes()
	assert.Equal(t, expected, result)
}

func TestWriteSingleRegisterRequest_Bytes(t *testing.T) {
	request := WriteSingleRegisterRequest{Offset: 0, Value: 0x0001}
	expected := []byte{0x00, 0x00, 0x00, 0x01}
	result := request.Bytes()
	assert.Equal(t, expected, result)
}

func TestWriteMultipleCoilsResponse_Bytes(t *testing.T) {
	response := WriteMultipleCoilsResponse{Offset: 0, Count: 16}
	expected := []byte{0x00, 0x00, 0x00, 0x10}
	result := response.Bytes()
	assert.Equal(t, expected, result)
}

func TestWriteMultipleCoilsRequest_Bytes(t *testing.T) {
	values := []bool{true, false, false, false, false, false, false, true, true, false, false, false, false, false, false, true}
	request := WriteMultipleCoilsRequest{Offset: 0, Values: values}
	expected := []byte{0x00, 0x00, 0x00, 0x10, 0x02, 0x81, 0x81}
	result := request.Bytes()
	assert.Equal(t, expected, result)
}

func TestWriteMultipleRegistersResponse_Bytes(t *testing.T) {
	response := WriteMultipleRegistersResponse{Offset: 0, Count: 4}
	expected := []byte{0x00, 0x00, 0x00, 0x04}
	result := response.Bytes()
	assert.Equal(t, expected, result)
}

func TestWriteMultipleRegistersRequest_Bytes(t *testing.T) {
	values := []uint16{0x0001, 0x0002, 0x0003, 0x0004}
	request := WriteMultipleRegistersRequest{Offset: 0, Values: values}
	expected := []byte{0x00, 0x00, 0x00, 0x04, 0x08, 0x00, 0x01, 0x00, 0x02, 0x00, 0x03, 0x00, 0x04}
	result := request.Bytes()
	assert.Equal(t, expected, result)
}

func TestMaskWriteRegisterResponse_Bytes(t *testing.T) {
	response := WriteMultipleRegistersResponse{Offset: 0, Count: 4}
	expected := []byte{0x00, 0x00, 0x00, 0x04}
	result := response.Bytes()
	assert.Equal(t, expected, result)
}
