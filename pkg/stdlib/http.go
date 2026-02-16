package stdlib

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/agenthands/nforth/pkg/core/value"
	"github.com/agenthands/nforth/pkg/vm"
)

var (
	ErrDomainNotAllowed = errors.New("stdlib/http: domain not allowed")
	ErrLocalhostBlocked = errors.New("stdlib/http: localhost/internal access blocked")
)

type HTTPSandbox struct {
	AllowedDomains []string
	AllowLocalhost bool
}

func NewHTTPSandbox(allowedDomains []string) *HTTPSandbox {
	return &HTTPSandbox{
		AllowedDomains: allowedDomains,
	}
}

// Fetch: ( url -- content )
func (s *HTTPSandbox) Fetch(m *vm.Machine) error {
	urlPacked := m.Pop().Data
	urlStr := value.UnpackString(urlPacked, m.Arena)

	u, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	// CONSTRAINT 1: Strict Allowlist
	if !s.isAllowed(u.Hostname()) {
		return ErrDomainNotAllowed
	}

	// CONSTRAINT 2: No Localhost
	if !s.AllowLocalhost && isLocalhost(u.Hostname()) {
		return ErrLocalhostBlocked
	}

	resp, err := http.Get(urlStr)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Push to Arena and Return
	offset := uint32(len(m.Arena))
	length := uint32(len(data))
	m.Arena = append(m.Arena, data...)

	m.Push(value.Value{
		Type: value.TypeString,
		Data: value.PackString(offset, length),
	})

	return nil
}

func (s *HTTPSandbox) isAllowed(hostname string) bool {
	for _, domain := range s.AllowedDomains {
		if hostname == domain || strings.HasSuffix(hostname, "."+domain) {
			return true
		}
	}
	return false
}

func isLocalhost(hostname string) bool {
	h := strings.ToLower(hostname)
	return h == "localhost" || h == "127.0.0.1" || h == "::1" || strings.HasPrefix(h, "192.168.") || strings.HasPrefix(h, "10.")
}
