package main_test

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSnippetComparison(t *testing.T) {
	snippets, err := filepath.Glob("snippets/*.py")
	if err != nil {
		t.Fatal(err)
	}

	if len(snippets) == 0 {
		t.Fatal("No snippets found")
	}

	npythonPath := "../npython"
	// Build npython first
	buildCmd := exec.Command("go", "build", "-o", npythonPath, "../cmd/npython")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build npython: %v\n%s", err, string(out))
	}

	failed := 0
	for _, snippet := range snippets {
		t.Run(snippet, func(t *testing.T) {
			// Run with Python3
			pyCmd := exec.Command("python3", snippet)
			pyOut, pyErr := pyCmd.CombinedOutput()

			// Run with nPython
			npyCmd := exec.Command(npythonPath, "run", snippet)
			npyOut, npyErr := npyCmd.CombinedOutput()

			// Compare errors
			if (pyErr == nil) != (npyErr == nil) {
				t.Errorf("Exit status mismatch for %s:\nPython3 err: %v\nnPython err: %v\nnPython out: %s", snippet, pyErr, npyErr, string(npyOut))
				failed++
				return
			}

			// Compare stdout
			sPyOut := strings.TrimSpace(string(pyOut))
			sNpyOut := strings.TrimSpace(string(npyOut))

			// Handle minor formatting differences in floats if necessary,
			// but for now, exact match of trimmed space.
			if sPyOut != sNpyOut {
				t.Errorf("Output mismatch for %s:\nPython3: [%s]\nnPython: [%s]", snippet, sPyOut, sNpyOut)
				failed++
			}
		})
	}

	fmt.Printf("\n--- Snippet Comparison Results ---\n")
	fmt.Printf("Total Snippets: %d\n", len(snippets))
	fmt.Printf("Passed: %d\n", len(snippets)-failed)
	fmt.Printf("Failed: %d\n", failed)

	if failed > 0 {
		t.Errorf("%d snippets failed", failed)
	}
}
