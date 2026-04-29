package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type LLMClient interface {
	Complete(ctx context.Context, prompt string, maxTokens int) (string, error)
}

type OllamaClient struct {
	http  *http.Client
	url   string
	model string
}

func NewOllamaClient() *OllamaClient {
	url := os.Getenv("OLLAMA_URL")
	if url == "" {
		url = "http://localhost:11434"
	}
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "llama3.2"
	}
	return &OllamaClient{
		http:  &http.Client{Timeout: 120 * time.Second},
		url:   url,
		model: model,
	}
}

func (c *OllamaClient) Complete(ctx context.Context, prompt string, _ int) (string, error) {
	type msg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	body, _ := json.Marshal(struct {
		Model    string `json:"model"`
		Messages []msg  `json:"messages"`
		Stream   bool   `json:"stream"`
	}{
		Model:    c.model,
		Messages: []msg{{Role: "user", Content: prompt}},
		Stream:   false,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama api: %w", err)
	}
	defer resp.Body.Close()

	var or struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&or); err != nil {
		return "", fmt.Errorf("ollama decode: %w", err)
	}
	if or.Error != "" {
		return "", fmt.Errorf("ollama error: %s", or.Error)
	}
	return or.Message.Content, nil
}
