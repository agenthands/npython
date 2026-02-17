package tests

import (
	"fmt"
	"strings"
	"testing"
)

type BenchmarkCase struct {
	Name        string
	Task        string
	NForthCode  string
	PythonCode  string
}

func countTokens(code string) int {
	// Approximation of LLM tokenization
	delimiters := []string{" ", "\n", "\t", "(", ")", "{", "}", ",", ":", ";", "\"", "'", "*", "/", "+", "-"}
	
	// Create a replacer to turn all delimiters into spaces
	replacements := make([]string, len(delimiters)*2)
	for i, d := range delimiters {
		replacements[i*2] = d
		replacements[i*2+1] = " "
	}
	r := strings.NewReplacer(replacements...)
	
	normalized := r.Replace(code)
	words := strings.Fields(normalized)
	return len(words)
}

func TestBenchmark_TokenEfficiency(t *testing.T) {
	cases := []BenchmarkCase{
		{
			Name: "Arithmetic & Tax",
			Task: "Calculate total price given base price and tax rate.",
			NForthCode: `: CALC-TOTAL { price tax }
  price tax MUL 100 DIV INTO amt
  price amt ADD INTO total
  total YIELD
;`,
			PythonCode: `def calc_total(price, tax):
    tax_amount = (price * tax) / 100
    return price + tax_amount`,
		},
		{
			Name: "Secure HTTP Fetch",
			Task: "Open HTTP gate, fetch URL, check for pattern, and return boolean.",
			NForthCode: `: CHECK-URL { url token }
  ADDRESS HTTP-ENV token
    url FETCH INTO html
  <EXIT>
  html "success" CONTAINS INTO ok
  ok YIELD
;`,
			PythonCode: `import requests
def check_url(url):
    try:
        resp = requests.get(url)
        return "success" in resp.text
    except:
        return False`,
		},
		{
			Name: "Secure File Write",
			Task: "Open FS gate, write data to path, and exit.",
			NForthCode: `: SAVE { data path token }
  ADDRESS FS-ENV token
    data path WRITE-FILE
  <EXIT>
;`,
			PythonCode: `import os
def save(data, path):
    # No built-in jailing in Python
    with open(path, "w") as f:
        f.write(data)`,
		},
	}

	fmt.Println("\nNFORTH VS PYTHON: TOKEN EFFICIENCY")
	fmt.Println("-------------------------------------------------------")
	fmt.Printf("%-20s %-10s %-10s %-10s\n", "Case", "nForth", "Python", "Saving %")
	fmt.Println("-------------------------------------------------------")

	for _, c := range cases {
		nfTokens := countTokens(c.NForthCode)
		pyTokens := countTokens(c.PythonCode)
		saving := 100.0 - (float64(nfTokens) / float64(pyTokens) * 100.0)
		
		fmt.Printf("%-20s %-10d %-10d %-10.1f%%\n", c.Name, nfTokens, pyTokens, saving)
	}
}
