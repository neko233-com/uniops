package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type CustomProvider struct {
	apiKey     string
	endpoint   string
	httpClient *http.Client
}

type CustomRequest struct {
	Messages []Message `json:"messages"`
	Model    string    `json:"model,omitempty"`
	Stream   bool      `json:"stream"`
}

type CustomResponse struct {
	Content string `json:"content"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func NewCustomProvider(apiKey, endpoint string) *CustomProvider {
	if endpoint == "" {
		endpoint = "http://localhost:11434/v1/chat/completions"
	}
	return &CustomProvider{
		apiKey:     apiKey,
		endpoint:   endpoint,
		httpClient: &http.Client{},
	}
}

func (p *CustomProvider) Name() string {
	return "custom"
}

func (p *CustomProvider) Chat(messages []Message) (string, error) {
	req := CustomRequest{
		Messages: messages,
		Stream:   false,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequest("POST", p.endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Try standard OpenAI-compatible format first
	var customResp CustomResponse
	if err := json.Unmarshal(respBody, &customResp); err == nil {
		if customResp.Content != "" {
			return customResp.Content, nil
		}
		if len(customResp.Choices) > 0 {
			return customResp.Choices[0].Message.Content, nil
		}
	}

	// Try plain text response
	return string(respBody), nil
}

func (p *CustomProvider) StreamChat(messages []Message) (<-chan string, error) {
	ch := make(chan string, 1)
	result, err := p.Chat(messages)
	if err != nil {
		close(ch)
		return nil, err
	}
	ch <- result
	close(ch)
	return ch, nil
}
