package value

import (
	"fmt"
	"math"
	"strings"
	"unsafe"
)

// Type represents the tag in the Value tagged union.
type Type uint8

const (
	TypeVoid Type = iota
	TypeInt
	TypeBool
	TypeFloat
	TypeString
	TypeBytes
	TypeDict
	TypeList
	TypeTuple
	TypeSet
	TypeIterator
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
	if !strings.Contains(s, "\\") {
		return s
	}
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

// Format returns a string representation of the value.
func (v Value) Format(arena []byte) string {
	switch v.Type {
	case TypeString:
		return UnpackString(v.Data, arena)
	case TypeInt:
		return strings.TrimSuffix(strings.TrimSuffix(fmt.Sprintf("%d", int64(v.Data)), ".0"), ".00")
	case TypeFloat:
		return fmt.Sprintf("%g", math.Float64frombits(v.Data))
	case TypeBool:
		if v.Data != 0 {
			return "True"
		}
		return "False"
	case TypeList, TypeTuple:
		var list []Value
		if v.Type == TypeList {
			if lp, ok := v.Opaque.(*[]Value); ok {
				list = *lp
			}
		} else {
			if l, ok := v.Opaque.([]Value); ok {
				list = l
			}
		}
		parts := make([]string, len(list))
		for i, el := range list {
			parts[i] = el.Format(arena)
		}
		if v.Type == TypeList {
			return "[" + strings.Join(parts, ", ") + "]"
		}
		return "(" + strings.Join(parts, ", ") + ")"
	case TypeDict:
		return fmt.Sprintf("%v", v.Opaque)
	case TypeVoid:
		return "None"
	default:
		return fmt.Sprintf("%v", v.Data)
	}
}
