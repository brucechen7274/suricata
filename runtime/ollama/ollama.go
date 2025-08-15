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

package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ostafen/suricata/runtime"
)

const DefaultBaseURL = "http://localhost:11434"

type OllamaInvoker struct {
	baseURL string
	model   string
	opts    Options
}

func NewInvoker(baseURL, model string, opts Options) *OllamaInvoker {
	return &OllamaInvoker{
		baseURL: baseURL,
		model:   model,
		opts:    opts,
	}
}

func roleToOllamaRole(role runtime.Role) string {
	switch role {
	case runtime.RoleSystem:
		return "system"
	case runtime.RoleAgent:
		return "assistant"
	case runtime.RoleUser:
		return "user"
	default:
		return "user"
	}
}

type OllamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Options struct {
	Temperature float64 `json:"temperature"`
	NumCtx      int     `json:"num_ctx"`
}

type OllamaPayload struct {
	Model    string          `json:"model"`
	Messages []OllamaMessage `json:"messages"`
	Prompt   string          `json:"prompt"`
	Stream   bool            `json:"stream"`
	Options  Options         `json:"options"`
}

func (o *OllamaInvoker) Invoke(ctx context.Context, systemPrompt string, messages []runtime.Message) (string, error) {
	payload := OllamaPayload{
		Model:    o.model,
		Messages: nil,
		Stream:   false,
		Options:  o.opts,
	}

	if systemPrompt != "" {
		payload.Messages = append(payload.Messages, OllamaMessage{
			Role:    roleToOllamaRole(runtime.RoleSystem),
			Content: systemPrompt,
		})
	}

	for _, m := range messages {
		payload.Messages = append(payload.Messages, OllamaMessage{
			Role:    roleToOllamaRole(m.Role),
			Content: m.Content,
		})
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/chat", o.baseURL), bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama error: %s", string(body))
	}

	var result struct {
		Message OllamaMessage `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Message.Content, nil
}
