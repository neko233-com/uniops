package agent

type Provider interface {
	Name() string
	Chat(messages []Message) (string, error)
	StreamChat(messages []Message) (<-chan string, error)
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type ChatResponse struct {
	Content string `json:"content"`
}
