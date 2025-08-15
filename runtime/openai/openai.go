// Copyright (c) 2025 Suricata Contributors
// Original Author: Stefano Scafiti
//
// This file is part of Suricata: Type-Safe AI Agents for Go.
//
// Licensed under the MIT License. You may obtain a copy of the License at
//
//	https://opensource.org/licenses/MIT
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package openai

import (
	"context"
	"errors"

	openai "github.com/sashabaranov/go-openai"
)

type Role uint8

const (
	RoleSystem Role = iota
	RoleAgent
	RoleUser
)

type Message struct {
	Role    Role
	Content string
}

// Invoker interface
type Invoker interface {
	Invoke(ctx context.Context, systemPrompt string, messages []Message) (string, error)
}

type OpenAIInvoker struct {
	client *openai.Client
	model  string
}

func NewInvoker(authToken string, model string) *OpenAIInvoker {
	return &OpenAIInvoker{
		client: openai.NewClient(authToken),
		model:  model,
	}
}

func roleToOpenAIRole(role Role) string {
	switch role {
	case RoleSystem:
		return "system"
	case RoleAgent:
		return "assistant"
	case RoleUser:
		return "user"
	default:
		return "user"
	}
}

func (o *OpenAIInvoker) Invoke(ctx context.Context, systemPrompt string, messages []Message) (string, error) {
	var chatMessages []openai.ChatCompletionMessage

	chatMessages = append(chatMessages, openai.ChatCompletionMessage{
		Role:    "system",
		Content: systemPrompt,
	})

	for _, m := range messages {
		chatMessages = append(chatMessages, openai.ChatCompletionMessage{
			Role:    roleToOpenAIRole(m.Role),
			Content: m.Content,
		})
	}

	resp, err := o.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    o.model,
		Messages: chatMessages,
	})
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("no response from OpenAI")
	}
	return resp.Choices[0].Message.Content, nil
}
