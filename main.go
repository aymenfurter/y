package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

const (
	MaxFileSize     = 5000
	OpenAIChatModel = "gpt-4"
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
	targetDirectory := "."
	err := filepath.Walk(targetDirectory, func(path string, info os.FileInfo, err error) error {
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
			fmt.Printf("Processing file: %s, line count: %d\n", filepath, len(lines))
			analyzeCodeWithChatBot(strings.Join(lines, "\n"), filepath)
			lines = nil
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
	}

	// Send remaining lines if any
	if len(lines) > 0 {
		fmt.Printf("Processing file: %s, line count: %d\n", filepath, len(lines))
		analyzeCodeWithChatBot(strings.Join(lines, "\n"), filepath)
	}
}

func analyzeCodeWithChatBot(code string, filepath string) {
	messages := []openai.ChatCompletionMessage{
		{
			Role:    "system",
			Content: "You are a pentest copilot assisting a security researcher identifying security vulnverabilities. Use OWASP best practice. Analyze the following code, think through it step-by-step. At the end of your analysis, give a security store from 1-10 in terms of severity. If there is none, state: 0\n\nExample:\n# Result\nScore: 0/10",
		},
		{
			Role:    "user",
			Content: code,
		},
	}

	_, err := chatWithopenaiWithMessages(messages, filepath)
	if err != nil {
		fmt.Printf("Failed to make request: %v\n", err)
	}
}

func chatWithopenaiWithMessages(messages []openai.ChatCompletionMessage, filepath string) (string, error) {
	ctx := context.Background()
	client := openai.NewClient(os.Getenv("OPEN_AI_KEY"))

	model := OpenAIChatModel

	fmt.Println("Scanning file: " + filepath)

	chatResp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    model,
		Messages: messages,
	})
	if err != nil {
		return "", err
	}

	// Extract from text "Score: 0" -> 0
	score := strings.Split(chatResp.Choices[0].Message.Content, "Score: ")[1]

	// check if score 9 or 10
	// check if score contains 9
	if strings.Contains(score, "9/10") || strings.Contains(score, "10/10") {
		fmt.Println("💣 This file has high severity: "+filepath, score)
		fmt.Println("Details: " + chatResp.Choices[0].Message.Content)

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Continue? (Y/N): ")
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		if text == "Y" {
			fmt.Println("Continuing...")
		}

	} else {
		fmt.Println("Completed " + filepath)
	}

	return chatResp.Choices[0].Message.Content, nil

}
