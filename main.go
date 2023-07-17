
package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	MaxFileSize     = 5000
	OpenAIChatModel = "gpt-4.0-turbo"
	OpenAIAPIKey    = "YOUR_OPENAI_API_KEY"
)

var (
	supportedLangs = map[string]bool{
		".go":   true,
		".py":   true,
		".js":   true,
		".java": true,
		// add more languages as needed
	}
	langCount = make(map[string]int)
)

func main() {
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		if !info.IsDir() {
			ext := filepath.Ext(path)
			if _, ok := supportedLangs[ext]; ok {
				langCount[ext]++
				processFile(path)
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("error walking the path: %v\n", err)
		return
	}

	// Print out language count
	for lang, count := range langCount {
		fmt.Printf("Files in %s: %d\n", lang, count)
	}
}

func processFile(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) == MaxFileSize {
			analyzeCodeWithChatBot(strings.Join(lines, "\n"))
			lines = nil
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
	}

	// Send remaining lines if any
	if len(lines) > 0 {
		analyzeCodeWithChatBot(strings.Join(lines, "\n"))
	}
}

func analyzeCodeWithChatBot(code string) {
	data := url.Values{}
	data.Set("model", OpenAIChatModel)
	data.Set("messages", fmt.Sprintf(`[{"role": "system", "content": "You are a helpful assistant."}, {"role": "user", "content": "%s"}]`, code))

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/engines/davinci/chat/completions", strings.NewReader(data.Encode()))
	if err != nil {
		fmt.Printf("Failed to create request: %v\n", err)
		return
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Bearer "+OpenAIAPIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Failed to make request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response body: %v\n", err)
		return
	}

	fmt.Println("ChatBot Response:", string(body))
}
