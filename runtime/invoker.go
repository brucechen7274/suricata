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

package runtime

import "context"

type Role uint8

const (
	RoleSystem = iota
	RoleAgent
	RoleUser
)

type Message struct {
	Role    Role
	Content string
}

// Invoker sends a prompt string to an LLM and returns the raw string response.
type Invoker interface {
	Invoke(ctx context.Context, systemPrompt string, messages []Message) (string, error)
}

type ChatSession struct {
	system   string
	messages []Message
	invoker  Invoker
}

func NewChatSession(invoker Invoker) *ChatSession {
	return &ChatSession{
		invoker:  invoker,
		messages: nil,
		system:   "", // TODO: integrate system prompt
	}
}

func (chat *ChatSession) Add(msg Message) {
	chat.messages = append(chat.messages, msg)
}

func (chat *ChatSession) Invoke(ctx context.Context, msg string) (string, error) {
	chat.Add(Message{Role: RoleUser, Content: msg})

	out, err := chat.invoker.Invoke(ctx, chat.system, chat.messages)
	if err != nil {
		return "", err
	}

	chat.Add(Message{Role: RoleAgent, Content: out})

	return out, nil
}
