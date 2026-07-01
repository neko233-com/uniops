package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type ClaudeProvider struct {
	apiKey     string
	endpoint   string
	model      string
	httpClient *http.Client
}

type ClaudeRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
}

type ClaudeResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

func NewClaudeProvider(apiKey, endpoint, model string) *ClaudeProvider {
	if endpoint == "" {
		endpoint = "https://api.anthropic.com/v1/messages"
	}
	if model == "" {
		model = "claude-3-5-sonnet-20241022"
	}
	return &ClaudeProvider{
		apiKey:     apiKey,
		endpoint:   endpoint,
		model:      model,
		httpClient: &http.Client{},
	}
}

func (p *ClaudeProvider) Name() string {
	return "claude"
}

func (p *ClaudeProvider) Chat(messages []Message) (string, error) {
	req := ClaudeRequest{
		Model:     p.model,
		MaxTokens: 4096,
		Messages:  messages,
	}

	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequest("POST", p.endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var claudeResp ClaudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return "", err
	}

	if len(claudeResp.Content) > 0 {
		return claudeResp.Content[0].Text, nil
	}

	return "", fmt.Errorf("no content in response")
}

func (p *ClaudeProvider) StreamChat(messages []Message) (<-chan string, error) {
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
