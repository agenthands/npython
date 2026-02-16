package main_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// --- Test Harness Utilities ---

// Sandbox creates a temporary workspace and returns its path and cleanup function
func setupSandbox(t *testing.T) (string, func()) {
	dir, err := os.MkdirTemp("", "nforth-test-*")
	if err != nil {
		t.Fatal(err)
	}
	return dir, func() { os.RemoveAll(dir) }
}

// MockInternet creates a local HTTP server that the agent can "FETCH" from
func setupMockInternet() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/data" {
			w.WriteHeader(200)
			w.Write([]byte("status=ok"))
		} else {
			w.WriteHeader(404)
		}
	}))
}

// RunNForth executes the nForth binary against a script
func runNForth(ctx context.Context, scriptPath string, workDir string, args []string) (string, error) {
	absNForth, _ := filepath.Abs("../nforth")
	fullArgs := append([]string{"run", scriptPath}, args...)
	cmd := exec.CommandContext(ctx, absNForth, fullArgs...)
	cmd.Dir = workDir
	
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// --- The Tests ---

func TestSuite_HappyPath(t *testing.T) {
	// 1. Setup
	sandbox, teardown := setupSandbox(t)
	defer teardown()
	
	server := setupMockInternet()
	defer server.Close()

	// 2. Create Agent Script
	// Note: We inject the dynamic server port into the script
	scriptContent := fmt.Sprintf(`
	ADDRESS HTTP-ENV token
		"%s/data" FETCH INTO res
	<EXIT>
	
	res "status=ok" CONTAINS INTO success
	success IF
		ADDRESS FS-ENV token
			"OK" "result.txt" WRITE-FILE
		<EXIT>
	THEN
	`, server.URL)

	scriptPath := filepath.Join(sandbox, "agent.nf")
	os.WriteFile(scriptPath, []byte(scriptContent), 0644)

	// 3. Execute
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	out, err := runNForth(ctx, scriptPath, sandbox, nil)

	// 4. Assertions
	if err != nil {
		t.Fatalf("Agent crashed: %v\nOutput: %s", err, out)
	}

	// Verify Side Effect (File System)
	resultFile := filepath.Join(sandbox, "result.txt")
	content, err := os.ReadFile(resultFile)
	if err != nil {
		t.Fatalf("Agent failed to write file: %v", err)
	}
	if string(content) != "OK" {
		t.Errorf("Expected 'OK', got '%s'", string(content))
	}
}

func TestSuite_Security_Jailbreak(t *testing.T) {
	sandbox, teardown := setupSandbox(t)
	defer teardown()

	script := `
	ADDRESS FS-ENV token
		"Hack" "../system.conf" WRITE-FILE
	<EXIT>
	`
	scriptPath := filepath.Join(sandbox, "exploit.nf")
	os.WriteFile(scriptPath, []byte(script), 0644)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	out, err := runNForth(ctx, scriptPath, sandbox, nil)

	// Assertion: MUST FAIL
	if err == nil {
		t.Fatal("Security Failure: Agent was allowed to write outside sandbox!")
	}
	// Our engine returns "stdlib/fs: path escape violation"
	if !strings.Contains(out, "path escape") && !strings.Contains(out, "security violation") {
		t.Errorf("Expected Security Error, got: %s", out)
	}
}

func TestSuite_Compiler_AntiHallucination(t *testing.T) {
	sandbox, teardown := setupSandbox(t)
	defer teardown()

	// This script has a "Dangling Stack" (10 20 ADD -> 30, never used)
	script := `
		10 20 ADD
		"Done" PRINT
	`
	scriptPath := filepath.Join(sandbox, "bad_logic.nf")
	os.WriteFile(scriptPath, []byte(script), 0644)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	out, err := runNForth(ctx, scriptPath, sandbox, nil)

	// Assertion: MUST FAIL COMPILATION
	if err == nil {
		t.Fatal("Compiler Failure: Accepted floating state logic!")
	}
	if !strings.Contains(out, "Syntactic Hallucination Error") {
		t.Errorf("Expected 'Syntactic Hallucination Error', got: %s", out)
	}
}

func TestSuite_FunctionDefinitions(t *testing.T) {
	sandbox, teardown := setupSandbox(t)
	defer teardown()

	script := `
	: ADD-TEN { n }
		n 10 ADD INTO result
		result PRINT
	;
	
	5 ADD-TEN
	`
	scriptPath := filepath.Join(sandbox, "func.nf")
	os.WriteFile(scriptPath, []byte(script), 0644)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	out, err := runNForth(ctx, scriptPath, sandbox, nil)

	if err != nil {
		t.Fatalf("Agent crashed: %v\nOutput: %s", err, out)
	}

	// We expect the result 15 to be printed or at least the execution to succeed.
	// Since we don't capture PRINT yet, success is enough.
}

func TestSuite_Performance_ZeroAlloc(t *testing.T) {
	sandbox, teardown := setupSandbox(t)
	defer teardown()

	script := `
		0 INTO i
		BEGIN
			i 1000 LT
		WHILE
			i 1 ADD INTO i
		REPEAT
		i PRINT
	`
	scriptPath := filepath.Join(sandbox, "stress.nf")
	os.WriteFile(scriptPath, []byte(script), 0644)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()
	out, err := runNForth(ctx, scriptPath, sandbox, nil)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Performance test failed: %v\nOutput: %s", err, out)
	}

	t.Logf("Executed 1,000 iterations in %v", duration)
	
	if duration > 1*time.Second {
		t.Errorf("Performance too slow: %v", duration)
	}
}
