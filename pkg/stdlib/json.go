package stdlib

import (
	"encoding/json"
	"fmt"
	"strings"
	"github.com/agenthands/nforth/pkg/core/value"
	"github.com/agenthands/nforth/pkg/vm"
)

// ParseJSON: ( str -- map )
func ParseJSON(m *vm.Machine) error {
	strPacked := m.Pop().Data
	str := value.UnpackString(strPacked, m.Arena)

	var data map[string]any
	if err := json.Unmarshal([]byte(str), &data); err != nil {
		return fmt.Errorf("json unmarshal failed: %v", err)
	}

	m.Push(value.Value{
		Type:   value.TypeMap,
		Opaque: data,
	})
	return nil
}

// ParseJSONKey: ( str key -- val )
func ParseJSONKey(m *vm.Machine) error {
	keyVal := m.Pop() // key is on top
	strVal := m.Pop() // str is below
	
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
	switch v := val.(type) {
	case string:
		offset := uint32(len(m.Arena))
		length := uint32(len(v))
		m.Arena = append(m.Arena, []byte(v)...)
		m.Push(value.Value{Type: value.TypeString, Data: value.PackString(offset, length)})
	case float64:
		m.Push(value.Value{Type: value.TypeInt, Data: uint64(int64(v))})
	case bool:
		var b uint64
		if v {
			b = 1
		}
		m.Push(value.Value{Type: value.TypeBool, Data: b})
	default:
		s := fmt.Sprintf("%v", v)
		offset := uint32(len(m.Arena))
		length := uint32(len(s))
		m.Arena = append(m.Arena, []byte(s)...)
		m.Push(value.Value{Type: value.TypeString, Data: value.PackString(offset, length)})
	}
	return nil
}

// GetField: ( map key -- val )
func GetField(m *vm.Machine) error {
	keyVal := m.Pop() // key is on top
	mapVal := m.Pop() // map is below
	
	key := value.UnpackString(keyVal.Data, m.Arena)
	key = strings.Trim(key, "\"") // Handle LLM quotes
	if mapVal.Type != value.TypeMap {
		return fmt.Errorf("expected Map, got %v", mapVal.Type)
	}

	data := mapVal.Opaque.(map[string]any)
	val, ok := data[key]
	if !ok {
		m.Push(value.Value{Type: value.TypeVoid})
		return nil
	}

	return pushConverted(m, val)
}
