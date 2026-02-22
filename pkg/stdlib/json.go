package stdlib

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/agenthands/npython/pkg/core/value"
	"github.com/agenthands/npython/pkg/vm"
)

// ParseJSON: ( str -- map )
func ParseJSON(m *vm.Machine) error {
	strVal := m.Pop()
	if strVal.Type != value.TypeString {
		return errors.New("TypeError: parse_json() argument 1 must be string")
	}
	str := value.UnpackString(strVal.Data, m.Arena)

	var data map[string]any
	if err := json.Unmarshal([]byte(str), &data); err != nil {
		return fmt.Errorf("json unmarshal failed: %v", err)
	}

	m.Push(value.Value{
		Type:   value.TypeDict,
		Opaque: data,
	})
	return nil
}

// ParseJSONKey: ( str key -- val )
func ParseJSONKey(m *vm.Machine) error {
	keyVal := m.Pop() // key is on top
	strVal := m.Pop() // str is below
	if keyVal.Type != value.TypeString || strVal.Type != value.TypeString {
		return errors.New("TypeError: parse_json_key() arguments must be strings")
	}

	key := value.UnpackString(keyVal.Data, m.Arena)
	str := value.UnpackString(strVal.Data, m.Arena)

	// Since we are doing a lot of unescaping in UnpackString,
	// let's ensure the string is a valid JSON before unmarshaling.
	// If it's wrapped in extra quotes, strip them.
	str = strings.Trim(str, "\"")
	str = strings.ReplaceAll(str, "\\\"", "\"")

	var data map[string]any
	if err := json.Unmarshal([]byte(str), &data); err != nil {
		return fmt.Errorf("json unmarshal failed: %v | Raw: %s", err, str)
	}

	val, ok := data[key]
	if !ok {
		m.Push(value.Value{Type: value.TypeVoid})
		return nil
	}

	return pushConverted(m, val)
}

func pushConverted(m *vm.Machine, val any) error {
	if v, ok := val.(value.Value); ok {
		m.Push(v)
		return nil
	}
	switch v := val.(type) {
	case string:
		offset, err := m.WriteArena([]byte(v))
		if err != nil {
			return err
		}
		m.Push(value.Value{Type: value.TypeString, Data: value.PackString(offset, uint32(len(v)))})
	case float64:
		m.Push(value.Value{Type: value.TypeInt, Data: uint64(int64(v))})
	case bool:
		var b uint64
		if v {
			b = 1
		}
		m.Push(value.Value{Type: value.TypeBool, Data: b})
	case map[string]any:
		m.Push(value.Value{Type: value.TypeDict, Opaque: v})
	case []any:
		results := make([]value.Value, len(v))
		for i, item := range v {
			if err := pushConverted(m, item); err != nil {
				return err
			}
			results[i] = m.Pop()
		}
		ptr := new([]value.Value)
		*ptr = results
		m.Push(value.Value{Type: value.TypeList, Opaque: ptr})
	case nil:
		m.Push(value.Value{Type: value.TypeVoid})
	default:
		s := fmt.Sprintf("%v", v)
		offset, err := m.WriteArena([]byte(s))
		if err != nil {
			return err
		}
		m.Push(value.Value{Type: value.TypeString, Data: value.PackString(offset, uint32(len(s)))})
	}
	return nil
}

// GetField: ( map key -- val )
func GetField(m *vm.Machine) error {
	keyVal := m.Pop() // key is on top
	mapVal := m.Pop() // map is below

	key := value.UnpackString(keyVal.Data, m.Arena)
	key = strings.Trim(key, "\"") // Handle LLM quotes
	if mapVal.Type != value.TypeDict {
		return fmt.Errorf("expected Dict, got %v", mapVal.Type)
	}

	data := mapVal.Opaque.(map[string]any)
	val, ok := data[key]
	if !ok {
		m.Push(value.Value{Type: value.TypeVoid})
		return nil
	}

	return pushConverted(m, val)
}
