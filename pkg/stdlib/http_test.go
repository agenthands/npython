package stdlib_test

import (
	"testing"
	"github.com/agenthands/npython/pkg/stdlib"
	"github.com/agenthands/npython/pkg/vm"
	"github.com/agenthands/npython/pkg/core/value"
)

func TestHTTPSandboxDomainBlocking(t *testing.T) {
	sandbox := stdlib.NewHTTPSandbox([]string{"google.com"})
	m := &vm.Machine{}
	
	// Setup Arena
	urlStr := "http://malicious.com"
	m.Arena = append(m.Arena, []byte(urlStr)...)
	urlData := value.PackString(0, uint32(len(urlStr)))

	m.Push(value.Value{Type: value.TypeString, Data: urlData})
	
	err := sandbox.Fetch(m)
	if err != stdlib.ErrDomainNotAllowed {
		t.Errorf("expected ErrDomainNotAllowed, got %v", err)
	}
}

func TestHTTPSandboxLocalhostBlocking(t *testing.T) {
	sandbox := stdlib.NewHTTPSandbox([]string{"localhost"})
	m := &vm.Machine{}
	
	// Setup Arena
	urlStr := "http://localhost:8080"
	m.Arena = append(m.Arena, []byte(urlStr)...)
	urlData := value.PackString(0, uint32(len(urlStr)))

	m.Push(value.Value{Type: value.TypeString, Data: urlData})
	
	err := sandbox.Fetch(m)
	if err != stdlib.ErrLocalhostBlocked {
		t.Errorf("expected ErrLocalhostBlocked, got %v", err)
	}
}
