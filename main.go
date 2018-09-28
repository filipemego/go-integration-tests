package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Test struct
type Test struct {
	Config Config
	Tests  []TestCase
}

// Config struct
type Config struct {
	BaseURL string
}

// TestCase struct
type TestCase struct {
	Group    string
	Name     string
	URL      string
	Method   string
	Headers  map[string]string
	Expected Expected
}

// Expected struct
type Expected struct {
	StatusCode int
}

func main() {
	var tests Test
	content, _ := ioutil.ReadFile("test.json")
	json.Unmarshal(content, &tests)

	runTests(tests)
}

func runTests(tests Test) {
	client := &http.Client{}

	for _, testCase := range tests.Tests {
		req, _ := http.NewRequest(testCase.Method, fmt.Sprintf("%s%s", tests.Config.BaseURL, testCase.URL), nil)

		setHeaders(req, testCase.Headers)

		resp, _ := client.Do(req)
		resp.Header
		assertExpects(resp, testCase.Expected)
	}
}

func setHeaders(req *http.Request, headers map[string]string) {
	if len(headers) > 0 {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}
}

func assertExpects(resp *http.Response, expected Expected) {
	if expected.StatusCode != 0 && expected.StatusCode != resp.StatusCode {
		fmt.Printf("status code, got %d, expected %d", resp.StatusCode, expected.StatusCode)
	}
}
