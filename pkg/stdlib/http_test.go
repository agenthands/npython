package stdlib_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/agenthands/npython/pkg/core/value"
	"github.com/agenthands/npython/pkg/stdlib"
	"github.com/agenthands/npython/pkg/vm"
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

func TestHTTPBuilder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/post" {
			w.WriteHeader(201)
			w.Write([]byte("created"))
		} else {
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	u, _ := url.Parse(server.URL)
	sandbox := stdlib.NewHTTPSandbox([]string{u.Hostname()})
	sandbox.AllowLocalhost = true
	m := vm.GetMachine()
	defer vm.PutMachine(m)
	m.Reset()

	if err := sandbox.WithClient(m); err != nil {
		t.Fatal(err)
	}

	m.Push(value.Value{Type: value.TypeString, Data: uint64(len(m.Arena))})
	m.Arena = append(m.Arena, server.URL+"/post"...)
	m.Push(value.Value{Type: value.TypeString, Data: value.PackString(uint32(len(m.Arena)-len(server.URL)-5), uint32(len(server.URL)+5))})
	if err := sandbox.SetURL(m); err != nil {
		t.Fatal(err)
	}

	m.Push(value.Value{Type: value.TypeString, Data: uint64(len(m.Arena))})
	m.Arena = append(m.Arena, "POST"...)
	m.Push(value.Value{Type: value.TypeString, Data: value.PackString(uint32(len(m.Arena)-4), 4)})
	if err := sandbox.SetMethod(m); err != nil {
		t.Fatal(err)
	}

	if err := sandbox.SendRequest(m); err != nil {
		t.Fatal(err)
	}
	resp := m.Pop()
	if resp.Type != value.TypeDict {
		t.Errorf("expected dict")
	}

	m.Push(resp)
	if err := sandbox.CheckStatus(m); err != nil {
		t.Fatal(err)
	}
	if m.Pop().Int() != 201 {
		t.Errorf("expected 201")
	}
}

func TestHTTPSandboxFetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	}))
	defer server.Close()

	u, _ := url.Parse(server.URL)
	sandbox := stdlib.NewHTTPSandbox([]string{u.Hostname()})
	sandbox.AllowLocalhost = true
	m := vm.GetMachine()
	defer vm.PutMachine(m)
	m.Reset()

	m.Arena = append(m.Arena, server.URL...)
	m.Push(value.Value{Type: value.TypeString, Data: value.PackString(0, uint32(len(server.URL)))})

	if err := sandbox.Fetch(m); err != nil {
		t.Fatal(err)
	}
	res := value.UnpackString(m.Pop().Data, m.Arena)
	if res != "hello" {
		t.Errorf("got %s", res)
	}
}
