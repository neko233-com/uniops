package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/neko233/uniops/internal/agent"
	"github.com/neko233/uniops/internal/deploy"
	"github.com/neko233/uniops/internal/store"
)

type AgentChatHandler struct {
	db *store.DB
}

type ChatRequest struct {
	AgentID  uint            `json:"agent_id"`
	ServerID uint            `json:"server_id"`
	Messages []agent.Message `json:"messages"`
}

type toolCall struct {
	Action string `json:"action"` // exec, deploy_nginx, deploy_backend, deploy_full
	Params struct {
		Command     string `json:"command"`
		ServiceName string `json:"service_name"`
		BinaryURL   string `json:"binary_url"`
		AppPort     int    `json:"app_port"`
		Domain      string `json:"domain"`
		NginxPort   int    `json:"nginx_port"`
	} `json:"params"`
}

var toolCallRegex = regexp.MustCompile("(?s)```tool\\s*\\n(.*?)\\n```")

const systemPrompt = `You are UniOps, a server operations AI agent. You manage Linux server clusters via SSH.

Available tools (output as code block with "tool" language tag):

1. Execute shell command on target server:
` + "```tool\n" + `{"action":"exec","params":{"command":"<shell command>"}}
` + "```" + `

2. Deploy nginx reverse proxy:
` + "```tool\n" + `{"action":"deploy_nginx","params":{"service_name":"<name>","app_port":<backend_port>,"domain":"<domain or _>","nginx_port":80}}
` + "```" + `

3. Deploy backend service:
` + "```tool\n" + `{"action":"deploy_backend","params":{"service_name":"<name>","binary_url":"<download URL>","app_port":<port>}}
` + "```" + `

4. Full deployment (backend + nginx):
` + "```tool\n" + `{"action":"deploy_full","params":{"service_name":"<name>","binary_url":"<URL>","app_port":<port>,"domain":"<domain>","nginx_port":80}}
` + "```" + `

Rules:
- Output ONE tool call per response when you need to execute something
- After executing, you'll receive the output/result
- Explain what you're doing and why before each tool call
- When done, summarize the results without a tool call
`

func NewAgentChatHandler(db *store.DB) *AgentChatHandler {
	return &AgentChatHandler{db: db}
}

func (h *AgentChatHandler) Chat(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	agentModel, err := h.db.GetAgent(req.AgentID)
	if err != nil {
		http.Error(w, "agent not found", http.StatusNotFound)
		return
	}

	var provider agent.Provider
	switch agentModel.Type {
	case "claude":
		provider = agent.NewClaudeProvider(agentModel.APIKey, agentModel.Endpoint, "")
	case "openai":
		provider = agent.NewOpenAIProvider(agentModel.APIKey, agentModel.Endpoint, "")
	case "custom":
		provider = agent.NewCustomProvider(agentModel.APIKey, agentModel.Endpoint)
	default:
		http.Error(w, "unsupported agent type", http.StatusBadRequest)
		return
	}

	// Inject system prompt
	messages := make([]agent.Message, 0, len(req.Messages)+1)
	messages = append(messages, agent.Message{Role: "system", Content: systemPrompt})

	if req.ServerID > 0 {
		server, err := h.db.GetServer(req.ServerID)
		if err == nil {
			messages = append(messages, agent.Message{
				Role:    "system",
				Content: fmt.Sprintf("Target server: %s (%s:%d, user: %s)", server.Name, server.Host, server.Port, server.Username),
			})
		}
	}

	messages = append(messages, req.Messages...)

	// Tool call loop: max 5 iterations
	deployer := deploy.NewService(h.db)
	for i := 0; i < 5; i++ {
		response, err := provider.Chat(messages)
		if err != nil {
			http.Error(w, "agent error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Check for tool calls
		matches := toolCallRegex.FindStringSubmatch(response)
		if matches == nil {
			// No tool call, return final response
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(agent.ChatResponse{Content: response})
			return
		}

		var call toolCall
		if err := json.Unmarshal([]byte(matches[1]), &call); err != nil {
			// Invalid tool call JSON, return response as-is
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(agent.ChatResponse{Content: response})
			return
		}

		// Execute the tool call
		result := h.executeToolCall(deployer, req.ServerID, call)

		// Add assistant response + tool result to conversation
		messages = append(messages, agent.Message{Role: "assistant", Content: response})
		messages = append(messages, agent.Message{
			Role:    "user",
			Content: fmt.Sprintf("Tool execution result:\n```\n%s\n```", result),
		})
	}

	// Max iterations reached, send last response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agent.ChatResponse{Content: "Max tool call iterations reached."})
}

func (h *AgentChatHandler) executeToolCall(deployer *deploy.Service, serverID uint, call toolCall) string {
	if serverID == 0 {
		return "Error: no target server specified"
	}

	server, err := h.db.GetServer(serverID)
	if err != nil {
		return fmt.Sprintf("Error: server not found: %v", err)
	}

	switch call.Action {
	case "exec":
		if call.Params.Command == "" {
			return "Error: no command specified"
		}
		output, err := deployer.ExecCommand(server, call.Params.Command)
		if err != nil {
			return fmt.Sprintf("Command failed: %v\nOutput:\n%s", err, output)
		}
		return output

	case "deploy_nginx":
		cfg := deploy.Config{
			ServiceName: valueOrDefault(call.Params.ServiceName, "uniops"),
			AppPort:     valueOrZero(call.Params.AppPort, 6020),
			Domain:      valueOrDefault(call.Params.Domain, "_"),
			NginxPort:   valueOrZero(call.Params.NginxPort, 80),
		}
		var logs strings.Builder
		err := deployer.DeployNginxStandalone(server, cfg, func(s string) {
			logs.WriteString(s + "\n")
		})
		if err != nil {
			return fmt.Sprintf("Deploy nginx failed: %v\n%s", err, logs.String())
		}
		return logs.String()

	case "deploy_backend":
		cfg := deploy.Config{
			ServiceName: valueOrDefault(call.Params.ServiceName, "uniops"),
			BinaryURL:   call.Params.BinaryURL,
			AppPort:     valueOrZero(call.Params.AppPort, 6020),
		}
		var logs strings.Builder
		err := deployer.DeployBackendStandalone(server, cfg, func(s string) {
			logs.WriteString(s + "\n")
		})
		if err != nil {
			return fmt.Sprintf("Deploy backend failed: %v\n%s", err, logs.String())
		}
		return logs.String()

	case "deploy_full":
		cfg := deploy.Config{
			ServiceName: valueOrDefault(call.Params.ServiceName, "uniops"),
			BinaryURL:   call.Params.BinaryURL,
			AppPort:     valueOrZero(call.Params.AppPort, 6020),
			Domain:      valueOrDefault(call.Params.Domain, "_"),
			NginxPort:   valueOrZero(call.Params.NginxPort, 80),
		}
		var logs strings.Builder
		logFn := func(s string) { logs.WriteString(s + "\n") }

		if err := deployer.DeployBackendStandalone(server, cfg, logFn); err != nil {
			return fmt.Sprintf("Deploy backend failed: %v\n%s", err, logs.String())
		}
		if err := deployer.DeployNginxStandalone(server, cfg, logFn); err != nil {
			return fmt.Sprintf("Deploy nginx failed: %v\n%s", err, logs.String())
		}
		return logs.String()

	default:
		return fmt.Sprintf("Unknown action: %s", call.Action)
	}
}

func valueOrDefault(val, def string) string {
	if val == "" {
		return def
	}
	return val
}

func valueOrZero(val, def int) int {
	if val == 0 {
		return def
	}
	return val
}
