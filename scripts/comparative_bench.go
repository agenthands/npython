package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Result struct {
	Task        string
	Environment string
	Status      string
	Error       string
	Tokens      int
	Duration    time.Duration
	Code        string
}

func main() {
	apiKey := os.Getenv("GEMINI_API_KEY")
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil { panic(err) }
	defer client.Close()

	model := client.GenerativeModel("gemini-2.0-flash")
	temp := float32(0.1)
	model.Temperature = &temp

	tasks := []string{
		"Fetch JSON from 'https://api.github.com/repos/google/go-github', extract the 'stargazers_count', and write it to a file named 'stars.txt'.",
		"Calculate the first 10 numbers of the Fibonacci sequence and print them as a list.",
		"Define a dictionary with 'name': 'agent', 'type': 'bot'. Update 'type' to 'ai-agent' and print the dictionary.",
		"Filter all words longer than 5 characters from a list of strings and print the result.",
	}

	promptData, _ := os.ReadFile("docs/PROMPT.md")
	datacardData, _ := os.ReadFile("docs/DATACARD.md")
	fullSystemPrompt := string(promptData) + "\n\n" + string(datacardData)

	fmt.Println("\nNPYTHON VS STANDARD PYTHON: AI AGENT BENCHMARK")
	fmt.Println("==================================================================")

	for i, task := range tasks {
		fmt.Printf("\nTASK %d: %s\n", i+1, task)
		
		pyRes := runTrial(ctx, model, task, "Standard Python", "Write standard Python 3 code. Use requests for HTTP. Assume requests is installed.", "")
		npyRes := runTrial(ctx, model, task, "nPython", "Write nPython code (secure subset). Use 'with scope(name, token)' for I/O. Tools: fetch(url), write_file(content, path), parse_json(str).", fullSystemPrompt)

		printComparison(pyRes, npyRes)
	}
}

func runTrial(ctx context.Context, model *genai.GenerativeModel, task, env, instruction, systemPrompt string) Result {
	if systemPrompt != "" {
		model.SystemInstruction = &genai.Content{
			Parts: []genai.Part{genai.Text(systemPrompt)},
		}
	} else {
		model.SystemInstruction = nil
	}

	prompt := fmt.Sprintf("Environment: %s\nConstraint: %s\nTask: %s\n\nOutput code ONLY inside a code block.", env, instruction, task)
	
	start := time.Now()
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return Result{Task: task, Environment: env, Status: "GEN_FAIL", Error: err.Error()}
	}
	
	code := extractCode(resp)
	tokens := countTokens(code)
	
	var status, errMsg string
	if env == "nPython" {
		status, errMsg = validateNPython(code)
	} else {
		status, errMsg = validatePython(code)
	}

	return Result{
		Task:        task,
		Environment: env,
		Status:      status,
		Error:       errMsg,
		Tokens:      tokens,
		Duration:    time.Since(start),
		Code:        code,
	}
}

func validateNPython(code string) (string, string) {
	tmp := "bench_check.py"
	os.WriteFile(tmp, []byte(code), 0644)
	defer os.Remove(tmp)

	cmd := exec.Command("./npython", "run", tmp)
	out, err := cmd.CombinedOutput()
	if err != nil {
		output := string(out)
		if strings.Contains(output, "security violation") {
			return "SEC_VIOLATION", "Attempted I/O without scope"
		}
		if strings.Contains(output, "Compilation Error") {
			return "HALLUCINATION", output
		}
		return "RUNTIME_ERR", output
	}
	return "PASS", ""
}

func validatePython(code string) (string, string) {
	tmp := "bench_check.py"
	os.WriteFile(tmp, []byte(code), 0644)
	defer os.Remove(tmp)

	cmd := exec.Command("python3", "-m", "py_compile", tmp)
	err := cmd.Run()
	if err != nil {
		return "SYNTAX_ERR", "Invalid Python syntax"
	}
	return "PASS", ""
}

func countTokens(s string) int {
	return len(strings.Fields(s))
}

func extractCode(resp *genai.GenerateContentResponse) string {
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 { return "" }
	text := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	if strings.Contains(text, "```python") {
		return strings.Split(strings.Split(text, "```python")[1], "```")[0]
	}
	return text
}

func printComparison(py, npy Result) {
	fmt.Printf("%-20s | %-15s | %-15s | %-10s\n", "Metric", "Std Python", "nPython", "Advantage")
	fmt.Println("-------------------------------------------------------------------------")
	fmt.Printf("%-20s | %-15s | %-15s | %-10s\n", "Status", py.Status, npy.Status, "")
	fmt.Printf("%-20s | %-15d | %-15d | %-10.1f%%\n", "Code Complexity (Tok)", py.Tokens, npy.Tokens, 100.0-(float64(npy.Tokens)/float64(py.Tokens)*100.0))
	
	if npy.Status != "PASS" && npy.Code != "" {
		fmt.Printf("nPython Code:\n%s\n", npy.Code)
	}
	if npy.Error != "" {
		shortErr := npy.Error
		if len(shortErr) > 120 { shortErr = shortErr[:117] + "..." }
		fmt.Printf("nPython Info: %s\n", strings.ReplaceAll(shortErr, "\n", " "))
	}
}
