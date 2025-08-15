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

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/xeipuuv/gojsonschema"
)

func TestRuntime_Invoke(t *testing.T) {
	type (
		Output struct {
			Result string `json:"result"`
		}
		Input struct {
			Name string `json:"name"`
		}
	)

	var (
		InputSchema  = gojsonschema.NewStringLoader(`{"type":"object","properties":{"name":{"type":"string"}},"required":["name"]}`)
		OutputSchema = gojsonschema.NewStringLoader(`{"type":"object","properties":{"result":{"type":"string"}},"required":["result"]}`)
	)

	t.Run("basic success no tools", func(t *testing.T) {
		mock := &mockInvoker{
			responses: []string{`{"result":"hello"}`},
		}

		rt := NewRuntime(mock)

		req := Request{
			PromptTemplate: "Hello, {{.Name}}",
			Input:          &Input{Name: "Pluto"},
			Output:         &Output{},
			InputSchema:    InputSchema,
			OutputSchema:   OutputSchema,
		}

		err := rt.Invoke(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		out := req.Output.(*Output)
		if out.Result != "hello" {
			t.Errorf("expected 'hello', got %q", out.Result)
		}
	})

	t.Run("invalid output JSON", func(t *testing.T) {
		mock := &mockInvoker{
			responses: []string{`not a json`},
		}
		rt := NewRuntime(mock)

		req := Request{
			PromptTemplate: "Hello",
			Input:          nil,
			Output:         &Output{},
			InputSchema:    InputSchema,
			OutputSchema:   OutputSchema,
		}

		err := rt.Invoke(context.Background(), req)
		if !errors.Is(err, ErrInvalidOutput) {
			t.Errorf("expected ErrInvalidOutput, got %v", err)
		}
	})

	t.Run("agent loop with tool call", func(t *testing.T) {
		mock := &mockInvoker{
			responses: []string{
				`{"name":"tool1","args":{"val":"x"},"done":false}`,
				`{"done":true,"out":{"result":"final"}}`,
			},
		}

		rt := NewRuntime(mock)

		toolCalled := false
		req := Request{
			PromptTemplate: "Tool test",
			Input:          &Input{},
			Output:         &Output{},
			InputSchema:    InputSchema,
			OutputSchema:   OutputSchema,
			ToolUnmarshaller: func(name string, data []byte) (any, error) {
				if name != "tool1" {
					return nil, fmt.Errorf("unexpected tool: %s", name)
				}
				var args struct{ Val string }
				_ = json.Unmarshal(data, &args)
				return args, nil
			},
			ToolInvoker: func(ctx context.Context, name string, in any) (any, error) {
				toolCalled = true
				return map[string]string{"toolResult": "ok"}, nil
			},
		}

		err := rt.Invoke(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !toolCalled {
			t.Errorf("tool was not invoked")
		}

		out := req.Output.(*Output)
		if out.Result != "final" {
			t.Errorf("expected 'final', got %q", out.Result)
		}
	})

	t.Run("context cancel in agent loop", func(t *testing.T) {
		mock := &mockInvoker{
			responses: []string{
				`{"name":"tool1","args":{"val":"x"},"done":false}`,
			},
		}

		rt := NewRuntime(mock)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel immediately

		req := Request{
			PromptTemplate: "Test",
			Input:          &Input{},
			Output:         &Output{},
			InputSchema:    InputSchema,
			OutputSchema:   OutputSchema,
			ToolUnmarshaller: func(name string, data []byte) (any, error) {
				return nil, nil
			},
			ToolInvoker: func(ctx context.Context, name string, in any) (any, error) {
				return nil, nil
			},
		}

		err := rt.Invoke(ctx, req)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	})
}

type mockInvoker struct {
	responses []string
	callCount int
}

func (m *mockInvoker) Invoke(ctx context.Context, input string, messages []Message) (string, error) {
	if m.callCount >= len(m.responses) {
		return "", fmt.Errorf("unexpected call")
	}
	resp := m.responses[m.callCount]
	m.callCount++
	return resp, nil
}
