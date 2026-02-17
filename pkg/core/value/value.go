package value

import (
	"strings"
	"unsafe"
)

// Type represents the tag in the Value tagged union.
type Type uint8

const (
	TypeVoid Type = iota
	TypeInt
	TypeBool
	TypeString
	TypeMap // For JSON objects
)

// Value is a tagged union.
type Value struct {
	Type   Type
	Data   uint64 // Still uint64 bits, but interpreted based on Type
	Opaque any    // For complex objects like maps
}

// PackString encodes offset and length into the Data register.
func PackString(offset, length uint32) uint64 {
	return (uint64(offset) << 32) | uint64(length)
}

// UnpackString retrieves a string view from the arena.
func UnpackString(data uint64, arena []byte) string {
	offset := uint32(data >> 32)
	length := uint32(data)

	if uint64(offset)+uint64(length) > uint64(len(arena)) {
		panic("value: memory access violation")
	}

	if length == 0 {
		return ""
	}

	s := unsafe.String(&arena[offset], length)
	s = strings.ReplaceAll(s, "\\\"", "\"")
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\t", "\t")
	s = strings.ReplaceAll(s, "\\r", "\r")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	return s
}

// Int returns the value as int64.
func (v Value) Int() int64 {
	return int64(v.Data)
}

// SetInt stores an int64.
func (v *Value) SetInt(i int64) {
	v.Type = TypeInt
	v.Data = uint64(i)
}
