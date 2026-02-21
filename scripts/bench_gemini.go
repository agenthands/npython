//go:build ignore

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

func main() {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: GEMINI_API_KEY not set")
		os.Exit(1)
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.0-flash")
	temp := float32(0.2)
	model.Temperature = &temp

	// Load System Prompt & Datacard
	prompt, _ := os.ReadFile("docs/PROMPT.md")
	datacard, _ := os.ReadFile("docs/DATACARD.md")

	fullSystemPrompt := string(prompt) + "\n\n" + string(datacard)
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(fullSystemPrompt)},
	}

	tasks := []string{
		"Calculate the factorial of 10 recursively and print the result. Do NOT wrap in a function definition unless you call it immediately.",
		"Calculate the square of each number in range(5) using map and lambda. Print the result list. Use top-level code.",
		"Filter even numbers from range(10) using filter and lambda, then sum them. Print the final sum. Use top-level code.",
		"Define an empty dictionary, set key 'info' to 'npython', and print the value of that key. Use top-level code.",
		"Create a list [10, 20, 30], change the middle element to 99, and print the entire list. Use top-level code.",
	}

	fmt.Println("\nBENCHMARKING NPYTHON WITH GEMINI FLASH")
	fmt.Println("--------------------------------------------------")

	for _, task := range tasks {
		fmt.Printf("Task: %s\n", task)
		start := time.Now()

		resp, err := model.GenerateContent(ctx, genai.Text(task))
		if err != nil {
			fmt.Printf("  GenAI Error: %v\n", err)
			continue
		}

		genDuration := time.Since(start)
		code := extractCode(resp)
		fmt.Printf("  Code:\n%s\n", code)

		// Save and Run
		os.WriteFile("bench_temp.py", []byte(code), 0644)

		runStart := time.Now()
		cmd := exec.Command("./npython", "run", "bench_temp.py")
		out, err := cmd.CombinedOutput()
		runDuration := time.Since(runStart)

		status := "PASS"
		if err != nil {
			status = "FAIL"
		}

		fmt.Printf("  Status: %s\n", status)
		fmt.Printf("  Gen Time: %v\n", genDuration)
		fmt.Printf("  Run Time: %v\n", runDuration)
		if status == "FAIL" {
			fmt.Printf("  Error: %s\n", string(out))
			fmt.Printf("  Code:\n%s\n", code)
		} else {
			fmt.Printf("  Output: %s", string(out))
		}
		fmt.Println("--------------------------------------------------")
	}
	os.Remove("bench_temp.py")
}

func extractCode(resp *genai.GenerateContentResponse) string {
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return ""
	}
	text := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	if strings.Contains(text, "```python") {
		return strings.Split(strings.Split(text, "```python")[1], "```")[0]
	}
	if strings.Contains(text, "```") {
		return strings.Split(strings.Split(text, "```")[1], "```")[0]
	}
	return text
}
