package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type OpenAIProvider struct {
	apiKey     string
	endpoint   string
	model      string
	httpClient *http.Client
}

type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func NewOpenAIProvider(apiKey, endpoint, model string) *OpenAIProvider {
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1/chat/completions"
	}
	if model == "" {
		model = "gpt-4"
	}
	return &OpenAIProvider{
		apiKey:     apiKey,
		endpoint:   endpoint,
		model:      model,
		httpClient: &http.Client{},
	}
}

func (p *OpenAIProvider) Name() string {
	return "openai"
}

func (p *OpenAIProvider) Chat(messages []Message) (string, error) {
	req := OpenAIRequest{
		Model:    p.model,
		Messages: messages,
	}

	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequest("POST", p.endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var openaiResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
		return "", err
	}

	if len(openaiResp.Choices) > 0 {
		return openaiResp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no content in response")
}

func (p *OpenAIProvider) StreamChat(messages []Message) (<-chan string, error) {
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
