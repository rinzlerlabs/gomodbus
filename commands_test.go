package gomodbus

import "testing"

func TestReadCoilsResponse_Bytes(t *testing.T) {
	values := []bool{false, true, false, true, false, false, false, false, true, false, false, false, true, false, false, false}
	response := ReadCoilsResponse{Values: values}
	expected := []byte{0x02, 0x0A, 0x11}
	result := response.Bytes()

	if len(result) != len(expected) {
		t.Fatalf("expected length %d, got %d", len(expected), len(result))
	}

	for i := range expected {
		if result[i] != expected[i] {
			t.Errorf("expected byte %d to be %02x, got %02x", i, expected[i], result[i])
		}
	}
}
