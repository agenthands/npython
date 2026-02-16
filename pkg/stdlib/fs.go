package stdlib

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/agenthands/nforth/pkg/core/value"
	"github.com/agenthands/nforth/pkg/vm"
)

var (
	ErrPathEscape     = errors.New("stdlib/fs: path escape violation")
	ErrFileTooLarge   = errors.New("stdlib/fs: file size limit exceeded")
	ErrPermissionDenied = errors.New("stdlib/fs: permission denied")
)

type FSSandbox struct {
	Root        string
	MaxFileSize int
}

func NewFSSandbox(root string, maxFileSize int) *FSSandbox {
	absRoot, _ := filepath.Abs(root)
	return &FSSandbox{
		Root:        absRoot,
		MaxFileSize: maxFileSize,
	}
}

// WriteFile: ( content path -- )
func (s *FSSandbox) WriteFile(m *vm.Machine) error {
	pathPacked := m.Pop().Data
	contentPacked := m.Pop().Data

	path := value.UnpackString(pathPacked, m.Arena)
	content := value.UnpackString(contentPacked, m.Arena)

	// CONSTRAINT A: Root Jailing
	cleanPath := filepath.Join(s.Root, filepath.Clean(path))
	if !strings.HasPrefix(cleanPath, s.Root) {
		return ErrPathEscape
	}

	// CONSTRAINT D: Size Limit
	if len(content) > s.MaxFileSize {
		return ErrFileTooLarge
	}

	// Create directory if not exists
	dir := filepath.Dir(cleanPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(cleanPath, []byte(content), 0644)
}

// ReadFile: ( path -- content )
func (s *FSSandbox) ReadFile(m *vm.Machine) error {
	pathPacked := m.Pop().Data
	path := value.UnpackString(pathPacked, m.Arena)

	// CONSTRAINT A: Root Jailing
	cleanPath := filepath.Join(s.Root, filepath.Clean(path))
	if !strings.HasPrefix(cleanPath, s.Root) {
		return ErrPathEscape
	}

	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return err
	}

	// We need to push the content to the Arena to return it as a string
	// For now, we'll append it to the Arena and push the packed value.
	offset := uint32(len(m.Arena))
	length := uint32(len(data))
	m.Arena = append(m.Arena, data...)

	m.Push(value.Value{
		Type: value.TypeString,
		Data: value.PackString(offset, length),
	})

	return nil
}
