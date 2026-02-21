package stdlib

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/agenthands/npython/pkg/core/value"
	"github.com/agenthands/npython/pkg/vm"
)

var (
	ErrDomainNotAllowed = errors.New("stdlib/http: domain not allowed")
	ErrLocalhostBlocked = errors.New("stdlib/http: localhost/internal access blocked")
)

type HTTPSandbox struct {
	AllowedDomains []string
	AllowLocalhost bool
	pendingReqs    map[*vm.Machine]*httpRequest
}

type httpRequest struct {
	url    string
	method string
	body   io.Reader
}

func NewHTTPSandbox(allowedDomains []string) *HTTPSandbox {
	return &HTTPSandbox{
		AllowedDomains: allowedDomains,
		pendingReqs:    make(map[*vm.Machine]*httpRequest),
	}
}

// WithClient: ( -- )
func (s *HTTPSandbox) WithClient(m *vm.Machine) error {
	s.pendingReqs[m] = &httpRequest{method: "GET"}
	return nil
}

// SetURL: ( url -- )
func (s *HTTPSandbox) SetURL(m *vm.Machine) error {
	urlPacked := m.Pop().Data
	urlStr := value.UnpackString(urlPacked, m.Arena)
	
	req, ok := s.pendingReqs[m]
	if !ok {
		return errors.New("stdlib/http: no pending request, call WITH-CLIENT first")
	}
	req.url = urlStr
	return nil
}

// SetMethod: ( method -- )
func (s *HTTPSandbox) SetMethod(m *vm.Machine) error {
	methodPacked := m.Pop().Data
	method := value.UnpackString(methodPacked, m.Arena)
	
	req, ok := s.pendingReqs[m]
	if !ok {
		return errors.New("stdlib/http: no pending request, call WITH-CLIENT first")
	}
	req.method = strings.ToUpper(method)
	return nil
}

// SendRequest: ( -- resp )
func (s *HTTPSandbox) SendRequest(m *vm.Machine) error {
	reqState, ok := s.pendingReqs[m]
	if !ok {
		return errors.New("stdlib/http: no pending request")
	}
	delete(s.pendingReqs, m)

	u, err := url.Parse(reqState.url)
	if err != nil {
		return err
	}

	if !s.isAllowed(u.Hostname()) {
		return ErrDomainNotAllowed
	}

	if !s.AllowLocalhost && isLocalhost(u.Hostname()) {
		return ErrLocalhostBlocked
	}

	httpClient := &http.Client{}
	req, err := http.NewRequest(reqState.method, reqState.url, reqState.body)
	if err != nil {
		return err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Push to Arena
	offset := uint32(len(m.Arena))
	length := uint32(len(data))
	m.Arena = append(m.Arena, data...)

	// Create a "Response" object. For now, we'll store status code in Opaque and body in String.
	// But the spec says 'resp CHECK-STATUS', so we need to return something that CHECK-STATUS can use.
	// We'll return a MAP or a special Response type.
	respMap := map[string]any{
		"status": int64(resp.StatusCode),
		"body":   value.Value{Type: value.TypeString, Data: value.PackString(offset, length)},
	}

	m.Push(value.Value{
		Type:   value.TypeDict,
		Opaque: respMap,
	})

	return nil
}

// CheckStatus: ( resp -- status )
func (s *HTTPSandbox) CheckStatus(m *vm.Machine) error {
	respVal := m.Pop()
	if respVal.Type != value.TypeDict {
		return errors.New("stdlib/http: CHECK-STATUS expects response map")
	}
	respMap := respVal.Opaque.(map[string]any)
	status := respMap["status"].(int64)
	
	m.Push(value.Value{Type: value.TypeInt, Data: uint64(status)})
	return nil
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
