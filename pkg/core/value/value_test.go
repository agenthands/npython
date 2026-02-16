package value_test

import (
	"testing"
	"github.com/agenthands/nforth/pkg/core/value"
)

func TestValueCreation(t *testing.T) {
	// Test Integer Value
	vInt := value.Value{Type: value.TypeInt, Data: 42}
	if vInt.Type != value.TypeInt {
		t.Errorf("expected TypeInt, got %v", vInt.Type)
	}
	if vInt.Data != 42 {
		t.Errorf("expected 42, got %v", vInt.Data)
	}

	// Test Boolean Value
	vBool := value.Value{Type: value.TypeBool, Data: 1}
	if vBool.Type != value.TypeBool {
		t.Errorf("expected TypeBool, got %v", vBool.Type)
	}
	if vBool.Data != 1 {
		t.Errorf("expected 1, got %v", vBool.Data)
	}
}

func TestStringPacking(t *testing.T) {
	arena := []byte("hello world")
	offset := uint32(0)
	length := uint32(5)

	data := value.PackString(offset, length)
	str := value.UnpackString(data, arena)

	if str != "hello" {
		t.Errorf("expected 'hello', got '%s'", str)
	}

	// Test another slice
	offset = 6
	length = 5
	data = value.PackString(offset, length)
	str = value.UnpackString(data, arena)

	if str != "world" {
		t.Errorf("expected 'world', got '%s'", str)
	}
}
