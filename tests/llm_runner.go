package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type BenchmarkCase struct {
	ID          string `json:"id"`
	Task        string `json:"task"`
	PromptRef   string `json:"prompt_ref"`
	ExpectedRef string `json:"expected_ref,omitempty"`
	Scoring     struct {
		Mode string `json:"mode"`
	} `json:"scoring"`
}

func main() {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
	}
	if apiKey == "" {
		fmt.Println("Error: GOOGLE_API_KEY or GEMINI_API_KEY not set")
		os.Exit(1)
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.0-flash")
	
	// Load System Prompt
	sysPrompt, _ := os.ReadFile("docs/PROMPT.md")
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(string(sysPrompt))},
	}

	// Paths to benchmark data
	basePath := "/Users/Janis_Vizulis/go/src/github.com/agenthands/anka-go-rlm-spec-v0.1/bench"
	casesDir := filepath.Join(basePath, "cases")

	suites, _ := filepath.Glob(filepath.Join(casesDir, "suite*.jsonl"))

	totalSuccess := 0
	totalCases := 0

	fmt.Println("\nNFORTH COMPREHENSIVE BENCHMARK")
	fmt.Println("-------------------------------------------------------")

	for _, suiteFile := range suites {
		suiteName := filepath.Base(suiteFile)
		fmt.Printf("\nSUITE: %s\n", suiteName)
		fmt.Println("-------------------------------------------------------")
		fmt.Printf("%-20s %-40s %-10s\n", "ID", "Task", "Status")
		fmt.Println("-------------------------------------------------------")

		f, err := os.Open(suiteFile)
		if err != nil {
			continue
		}
		
		decoder := json.NewDecoder(f)
		for decoder.More() {
			var bc BenchmarkCase
			if err := decoder.Decode(&bc); err != nil {
				break
			}
			totalCases++

			// Load Task Prompt
			promptPath := filepath.Join(basePath, bc.PromptRef)
			promptContent, _ := os.ReadFile(promptPath)

			query := fmt.Sprintf(`Task: %s

STRICT RULES:
1. Output TOP-LEVEL code only. DO NOT wrap in a function.
2. If the task requires processing the data provided below, follow the nForth paradigm:
   - Define data: "<DATA>" INTO input-data
   - Extract if JSON: input-data PARSE-JSON INTO data-map, then data-map "key" EXTRACT-KEY INTO result
   - Print result: result PRINT
3. If the task is about specific literals or logic, implement it directly.

Data provided:
"""
%s
"""
`, bc.Task, string(promptContent))

			resp, err := model.GenerateContent(ctx, genai.Text(query))
			if err != nil {
				fmt.Printf("%-20s %-40s %-10s\n", bc.ID, bc.Task, "ERROR (GenAI)")
				continue
			}

			rawOutput := extractCode(resp)
			passed, actual := runAndVerify(rawOutput, string(promptContent), bc, basePath)
			
			status := "FAILED"
			if passed {
				status = "PASSED"
				totalSuccess++
			}

			fmt.Printf("%-20s %-40s %-10s\n", bc.ID, bc.Task, status)
			if !passed {
				shortActual := strings.ReplaceAll(actual, "\n", " ")
				if len(shortActual) > 60 {
					shortActual = shortActual[:57] + "..."
				}
				fmt.Printf("    Actual: %s\n", shortActual)
			}
		}
		f.Close()
	}

	fmt.Println("\n=======================================================")
	if totalCases > 0 {
		fmt.Printf("GRAND TOTAL: %d/%d (%.1f%% success)\n", totalSuccess, totalCases, float64(totalSuccess)/float64(totalCases)*100)
	}
	fmt.Println("=======================================================")
}

func extractCode(resp *genai.GenerateContentResponse) string {
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return ""
	}
	part := resp.Candidates[0].Content.Parts[0]
	text := fmt.Sprintf("%v", part)
	
	if strings.Contains(text, "</thinking>") {
		text = strings.Split(text, "</thinking>")[1]
	}

	if strings.Contains(text, "```forth") {
		return strings.Split(strings.Split(text, "```forth")[1], "```")[0]
	}
	if strings.Contains(text, "```") {
		return strings.Split(strings.Split(text, "```")[1], "```")[0]
	}
	return text
}

func runAndVerify(code, inputData string, bc BenchmarkCase, basePath string) (bool, string) {
	tmpFile := "bench_tmp.nf"
	os.WriteFile(tmpFile, []byte(code), 0644)
	defer os.Remove(tmpFile)

	cmd := exec.Command("./nforth", "run", tmpFile)
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if bc.Scoring.Mode == "status_compile_error" {
		return err != nil && (strings.Contains(outputStr, "Compilation Error") || strings.Contains(outputStr, "Syntactic Hallucination")), outputStr
	}

	if err != nil {
		return false, outputStr
	}

	if bc.Scoring.Mode == "status_ok" {
		return true, outputStr
	}

	if bc.ExpectedRef == "" {
		return true, outputStr
	}

	// Load expected output
	expPath := filepath.Join(basePath, bc.ExpectedRef)
	expData, _ := os.ReadFile(expPath)
	expected := strings.TrimSpace(string(expData))
	
	// JSON semantic check
	var m map[string]any
	if err := json.Unmarshal(expData, &m); err == nil {
		// If expected IS a JSON object, check all fields
		for _, v := range m {
			vStr := fmt.Sprintf("%v", v)
			if !strings.Contains(outputStr, vStr) {
				return false, outputStr
			}
		}
		return true, outputStr
	}

	// Simple substring check for scalar results (numbers, etc)
	return strings.Contains(outputStr, expected), outputStr
}
