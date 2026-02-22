package stdlib

import (
	"fmt"
	"strings"

	"github.com/agenthands/npython/pkg/core/value"
	"github.com/agenthands/npython/pkg/vm"
)

// FormatString: ( format val -- result )
func FormatString(m *vm.Machine) error {
	valVal := m.Pop()
	formatVal := m.Pop()

	if formatVal.Type != value.TypeString {
		return fmt.Errorf("TypeError: format must be string")
	}
	format := value.UnpackString(formatVal.Data, m.Arena)

	var valStr string
	if valVal.Type == value.TypeString {
		valStr = value.UnpackString(valVal.Data, m.Arena)
	} else if valVal.Type == value.TypeInt {
		valStr = fmt.Sprintf("%d", int64(valVal.Data))
	} else if valVal.Type == value.TypeBool {
		valStr = fmt.Sprintf("%v", valVal.Data != 0)
	} else {
		valStr = fmt.Sprintf("%v", valVal.Data)
	}

	result := strings.Replace(format, "%s", valStr, 1)

	offset, err := m.WriteArena([]byte(result))
	if err != nil {
		return err
	}

	m.Push(value.Value{
		Type: value.TypeString,
		Data: value.PackString(offset, uint32(len(result))),
	})

	return nil
}

// IsEmpty: ( val -- bool )
func IsEmpty(m *vm.Machine) error {
	val := m.Pop()
	var empty bool
	if val.Type == value.TypeString {
		empty = val.Data == 0 || uint32(val.Data) == 0
	} else if val.Type == value.TypeVoid {
		empty = true
	} else if val.Type == value.TypeDict {
		if m, ok := val.Opaque.(map[string]any); ok {
			empty = len(m) == 0
		}
	} else {
		empty = val.Data == 0
	}

	var res uint64
	if empty {
		res = 1
	}
	m.Push(value.Value{Type: value.TypeBool, Data: res})
	return nil
}
