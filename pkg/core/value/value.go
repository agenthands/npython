package value

import (
	"unsafe"
)

// Type represents the tag in the Value tagged union.
type Type uint8

const (
	TypeVoid Type = iota
	TypeInt
	TypeBool
	TypeString
)

// Value is a 16-byte tagged union (Type + 64-bit Data).
// It is passed by value, never by pointer, to ensure zero-allocation.
type Value struct {
	Type Type
	Data uint64
}

// PackString encodes offset and length into the Data register.
// Layout: [32-bit Offset | 32-bit Length]
func PackString(offset, length uint32) uint64 {
	return (uint64(offset) << 32) | uint64(length)
}

// UnpackString retrieves a string view from the arena without allocation.
// It uses unsafe.String for maximum performance.
func UnpackString(data uint64, arena []byte) string {
	offset := uint32(data >> 32)
	length := uint32(data)

	// Minimalist bounds check
	if uint64(offset)+uint64(length) > uint64(len(arena)) {
		panic("value: memory access violation")
	}

	if length == 0 {
		return ""
	}

	return unsafe.String(&arena[offset], length)
}
