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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

var ErrInvalidOutput = errors.New("invalid output")

type (
	ToolUnmarshaller func(name string, data []byte) (any, error)
	ToolInvoker      func(ctx context.Context, name string, in any) (any, error)

	ToolSpec struct {
		Name        string
		Description string
		Schema      gojsonschema.JSONLoader
	}

	ToolResponse struct {
		Done bool `json:"done"`
		Out  any  `json:"out"`

		Name string `json:"name"`
		Args any    `json:"args"`
	}

	Request struct {
		SkipInput      bool
		Instructions   string
		PromptTemplate string // Go template string for the prompt
		Input          any    // Data passed to the prompt template
		Output         any
		InputSchema    gojsonschema.JSONLoader
		OutputSchema   gojsonschema.JSONLoader // Pointer to struct to unmarshal output JSON into

		ToolUnmarshaller ToolUnmarshaller
		ToolInvoker      ToolInvoker
		ToolSpecs        []ToolSpec
	}

	Runtime struct {
		invoker Invoker
	}
)

func NewRuntime(invoker Invoker) *Runtime {
	return &Runtime{
		invoker: invoker,
	}
}

func (r *Runtime) Invoke(ctx context.Context, req Request) error {
	if err := ValidateJSON(req.Input, req.InputSchema); err != nil {
		return err
	}

	prompt, err := r.preparePrompt(&req)
	if err != nil {
		return err
	}

	sess := NewChatSession(r.invoker, req.Instructions)

	out, err := sess.Invoke(
		ctx,
		prompt,
	)
	if err != nil {
		return err
	}

	if req.ToolInvoker == nil {
		return unmarshalOutput(out, &req)
	}
	return r.agentLoop(ctx, out, &req, sess)
}

func (r *Runtime) agentLoop(ctx context.Context, out string, req *Request, sess *ChatSession) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		resp, err := parseToolResponse(out)
		if err != nil {
			return err
		}

		if resp.Done {
			rawOut, err := json.Marshal(resp.Out)
			if err != nil {
				return fmt.Errorf("marshal final output: %w", err)
			}
			return unmarshalOutput(string(rawOut), req)
		}

		// Validate tool name and args
		if resp.Name == "" {
			return errors.New("tool response missing 'name'")
		}
		if resp.Args == nil {
			return fmt.Errorf("tool '%s' missing 'args'", resp.Name)
		}

		// Convert raw args into typed input
		rawArgs, err := json.Marshal(resp.Args)
		if err != nil {
			return fmt.Errorf("marshal tool args: %w", err)
		}

		inType, err := req.ToolUnmarshaller(resp.Name, rawArgs)
		if err != nil {
			return fmt.Errorf("tool unmarshal for '%s': %w", resp.Name, err)
		}

		toolOutput := r.callTool(ctx, resp.Name, inType, req)

		out, err = sess.Invoke(ctx, toolOutput)
		if err != nil {
			return fmt.Errorf("invoke session after tool '%s': %w", resp.Name, err)
		}
	}
}

func parseToolResponse(raw string) (ToolResponse, error) {
	rawJSON := ExtractJSONFromString(raw)
	if rawJSON == "" {
		return ToolResponse{}, errors.New("no valid JSON found in response")
	}

	var resp ToolResponse
	if err := json.Unmarshal([]byte(rawJSON), &resp); err != nil {
		return ToolResponse{}, fmt.Errorf("invalid JSON format: %s", rawJSON)
	}
	return resp, nil
}

func (r *Runtime) callTool(ctx context.Context, name string, inType any, req *Request) string {
	toolResp, err := req.ToolInvoker(ctx, name, inType)
	if err != nil {
		return "ERR: " + err.Error()
	}

	rawToolResp, _ := json.Marshal(toolResp)
	return name + " OUTPUT: " + string(rawToolResp)
}

func unmarshalOutput(out string, req *Request) error {
	out = ExtractJSONFromString(out)
	if out == "" {
		return ErrInvalidOutput
	}
	return UnmarshalValidate([]byte(out), req.Output, req.OutputSchema)
}

func (r *Runtime) preparePrompt(req *Request) (string, error) {
	compiledPrompt, err := r.compilePrompt(req)
	if err != nil {
		return "", err
	}

	var pb PromptBuilder

	prompt := pb.Build(compiledPrompt, req)
	return prompt, nil
}

func (r *Runtime) compilePrompt(req *Request) (string, error) {
	// TODO: add more utility functions
	funcMap := template.FuncMap{
		"join": strings.Join,
	}

	tmpl, err := template.New("prompt").
		Funcs(funcMap).
		Parse(req.PromptTemplate)
	if err != nil {
		return "", fmt.Errorf("template parse: %w", err)
	}

	var promptBuf bytes.Buffer
	if err := tmpl.Execute(&promptBuf, req.Input); err != nil {
		return "", fmt.Errorf("template execute: %w", err)
	}

	prompt := promptBuf.String()
	return prompt, nil
}
