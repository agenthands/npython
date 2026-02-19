package stdlib_test

import (
	"os"
	"path/filepath"
	"testing"
	"github.com/agenthands/npython/pkg/stdlib"
	"github.com/agenthands/npython/pkg/vm"
	"github.com/agenthands/npython/pkg/core/value"
)

func TestFSSandbox(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "npython-fs-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sandbox := stdlib.NewFSSandbox(tempDir, 1024)
	m := &vm.Machine{}
	
	// Setup Arena for Write
	pathStr := "test.txt"
	contentStr := "hello npython"
	
	m.Arena = append(m.Arena, []byte(contentStr)...)
	contentData := value.PackString(0, uint32(len(contentStr)))
	
	pathOffset := uint32(len(m.Arena))
	m.Arena = append(m.Arena, []byte(pathStr)...)
	pathData := value.PackString(pathOffset, uint32(len(pathStr)))

	// Test Write ( content path -- )
	m.Push(value.Value{Type: value.TypeString, Data: contentData})
	m.Push(value.Value{Type: value.TypeString, Data: pathData})
	
	err = sandbox.WriteFile(m)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filepath.Join(tempDir, pathStr)); os.IsNotExist(err) {
		t.Errorf("file test.txt was not created")
	}

	// Test Read ( path -- content )
	m.Reset()
	m.Arena = m.Arena[:0] // Clear arena
	
	m.Arena = append(m.Arena, []byte(pathStr)...)
	pathData = value.PackString(0, uint32(len(pathStr)))
	m.Push(value.Value{Type: value.TypeString, Data: pathData})

	err = sandbox.ReadFile(m)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	res := m.Pop()
	resStr := value.UnpackString(res.Data, m.Arena)
	if resStr != contentStr {
		t.Errorf("expected '%s', got '%s'", contentStr, resStr)
	}
}

func TestFSSandboxJailing(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "npython-fs-jail-test")
	defer os.RemoveAll(tempDir)
	
	sandbox := stdlib.NewFSSandbox(tempDir, 1024)
	m := &vm.Machine{}

	// Attempt path escape ( "../../etc/passwd" )
	pathStr := "../../etc/passwd"
	contentStr := "malicious content"
	
	m.Arena = append(m.Arena, []byte(contentStr)...)
	contentData := value.PackString(0, uint32(len(contentStr)))
	
	pathOffset := uint32(len(m.Arena))
	m.Arena = append(m.Arena, []byte(pathStr)...)
	pathData := value.PackString(pathOffset, uint32(len(pathStr)))

	m.Push(value.Value{Type: value.TypeString, Data: contentData})
	m.Push(value.Value{Type: value.TypeString, Data: pathData})

	err := sandbox.WriteFile(m)
	if err != stdlib.ErrPathEscape {
		t.Errorf("expected ErrPathEscape, got %v", err)
	}
}
