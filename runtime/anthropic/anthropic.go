package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ostafen/suricata/runtime"
)

type Model string

const (
	AnthropicBaseURL = "https://api.anthropic.com/v1/messages"
	AnthropicVersion = "2023-06-01"

	ClaudeSonnet35V2 Model = "claude-3-5-sonnet-20241022"
	ClaudeSonnet37   Model = "claude-3-7-sonnet-20250219"
	ClaudeOpus4      Model = "claude-opus-4-20250514"
	ClaudeSonnet4    Model = "claude-sonnet-4-20250514"
)

const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleSystem    = "system"
)

// Message represents a single message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AnthropicInvoker is the client for Anthropic API
type AnthropicInvoker struct {
	APIKey    string
	Model     Model
	MaxTokens int
}

// NewAnthropicInvoker creates a new invoker instance
func NewAnthropicInvoker(apiKey string, model Model, maxTokens int) *AnthropicInvoker {
	return &AnthropicInvoker{
		APIKey:    apiKey,
		Model:     model,
		MaxTokens: maxTokens,
	}
}

// anthropicRequest represents the request payload
type anthropicRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
}

// anthropicResponse represents the response from Anthropic API
type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

// Invoke sends a set of messages and returns the assistant response
func (a *AnthropicInvoker) Invoke(ctx context.Context, system string, messages []runtime.Message) (string, error) {
	reqBody := anthropicRequest{
		Model:     string(a.Model),
		MaxTokens: a.MaxTokens,
		Messages:  toAnthropicMessages(messages),
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, AnthropicBaseURL, bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-api-key", a.APIKey)
	req.Header.Set("content-type", "application/json")
	req.Header.Set("anthropic-version", AnthropicVersion)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("non-200 status: %d, body: %s", resp.StatusCode, body)
	}

	var anthropicResp anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Combine all text parts
	var result string
	for _, c := range anthropicResp.Content {
		if c.Type == "text" {
			result += c.Text
		}
	}
	return result, nil
}

func toAnthropicMessages(messages []runtime.Message) []Message {
	out := make([]Message, len(messages))

	for i, msg := range messages {
		out[i] = Message{
			Role:    getRole(msg.Role),
			Content: msg.Content,
		}
	}
	return out
}

func getRole(r runtime.Role) string {
	switch r {
	case runtime.RoleAgent:
		return RoleAssistant
	case runtime.RoleUser:
		return RoleUser
	}
	return ""
}
