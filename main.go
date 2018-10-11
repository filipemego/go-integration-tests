package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"gopkg.in/yaml.v2"
)

// Headers type
type Headers map[string]string

// Root struct
type Root struct {
	Config Config `yaml:"config"`
	Tests  []Test `yaml:"tests"`
}

// Config struct
type Config struct {
	BaseURL string `yaml:"baseUrl"`
	Timeout int    `yaml:"timeout"`
}

// Test struct
type Test struct {
	Group    string     `yaml:"group"`
	Name     string     `yaml:"name"`
	URL      string     `yaml:"url"`
	Method   string     `yaml:"method"`
	Headers  Headers    `yaml:"headers"`
	Body     string     `yaml:"body"`
	Expected []Expected `yaml:"expected"`
}

// Expected struct
type Expected struct {
	StatusCode int     `yaml:"statusCode"`
	Headers    Headers `yaml:"headers"`
	Body       string  `yaml:"body"`
}

func main() {
	workdir := flag.String("dir", "", "directory of the tests files")
	flag.Parse()
	println(*workdir)

	files, err := ioutil.ReadDir(*workdir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), "yaml") || strings.HasSuffix(file.Name(), "yml") {
			content, _ := ioutil.ReadFile(fmt.Sprintf("%s/%s", *workdir, file.Name()))
			root := fromYAML(content)
			runTests(file.Name(), root)
			color("reset")
			fmt.Printf("\n-----------------------------------------\n\n")
		}
	}
}

func fromYAML(content []byte) *Root {
	var tests Root
	yaml.Unmarshal(content, &tests)
	return &tests
}

func runTests(filename string, root *Root) {
	client := &http.Client{}

	errors := make(map[int][]error)
	total := len(root.Tests)
	totalOK := 0

	fmt.Printf("running tests for %s\n", filename)
	for index, testCase := range root.Tests {
		var req *http.Request
		req, _ = http.NewRequest(testCase.Method, fmt.Sprintf("%s%s", root.Config.BaseURL, testCase.URL), strings.NewReader(testCase.Body))
		setHeaders(req, testCase.Headers)

		resp, _ := client.Do(req)
		if resp != nil {
			defer resp.Body.Close()
		}

		testErr := assertExpected(resp, testCase.Expected)
		if len(testErr) != 0 {
			errors[index] = testErr
		} else {
			totalOK++
		}

	}

	if errors != nil {
		printErrors(root.Tests, errors)
		return
	}

	color("green")
	fmt.Printf("\033[0;32mtest.yaml - %d/%d passed\n", totalOK, total)
}

func setHeaders(req *http.Request, headers map[string]string) {
	if len(headers) > 0 {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}
}

func assertExpected(resp *http.Response, expects []Expected) []error {
	var errors []error

	if resp != nil {
		for _, expected := range expects {
			if scErr := assertStatusCode(resp.StatusCode, expected.StatusCode); scErr != nil {
				errors = append(errors, scErr)
				continue
			}

			if hErr := assertHeaders(resp.Header, expected.Headers); hErr != nil {
				errors = append(errors, hErr)
				continue
			}

			body, rbErr := ioutil.ReadAll(resp.Body)
			if rbErr == nil {
				if bErr := assetFullBody(string(body), expected.Body); bErr != nil {
					errors = append(errors, bErr)
				}
			}
		}
	}
	return errors
}

func assetFullBody(got, expected string) error {
	fmt.Printf("===>%s\n", got)
	if expected != "" && expected != got {
		return fmt.Errorf("body: got %s, expected %s", got, expected)
	}
	return nil
}

func assertStatusCode(got, expected int) error {
	if expected != 0 && expected != got {
		return fmt.Errorf("status code: got %d, expected %d", got, expected)
	}
	return nil
}

func assertHeaders(got http.Header, expected Headers) error {
	if expected != nil {
		for header, value := range expected {
			for gotHeader, gotHValues := range got {
				if header == gotHeader {
					for _, gotValue := range gotHValues {
						if !strings.Contains(gotValue, value) {
							return fmt.Errorf("header: got %s, expected %s", gotValue, value)
						}
					}
				}
			}
		}
	}
	return nil
}

func printErrors(tests []Test, errors map[int][]error) {
	for index, testErrs := range errors {
		color("red")
		fmt.Printf("\n[%s] got %d error(s):\n", tests[index].Name, len(testErrs))

		for _, err := range testErrs {
			fmt.Printf("\t- %s\n", err)
		}
	}
}

var colors = map[string]string{
	"green":  "\033[32m",
	"red":    "\033[31m",
	"yellow": "\033[33m",
	"reset":  "\033[0m",
}

func color(color string) {
	fmt.Printf("%s", colors[color])
}
