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
package runtime_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ostafen/suricata/runtime"
	"github.com/xeipuuv/gojsonschema"
)

func TestPromptBuilder_Build(t *testing.T) {
	// Arrange
	inputData := map[string]string{"query": "test search"}
	outputSchema := gojsonschema.NewStringLoader(`{
		"type": "object",
		"properties": {
			"done": {"type": "boolean"},
			"out": {"type": "object"}
		},
		"required": ["done", "out"]
	}`)

	toolSchema := gojsonschema.NewStringLoader(`{
		"type": "object",
		"properties": {"query": {"type": "string"}},
		"required": ["query"]
	}`)

	tools := []runtime.ToolSpec{
		{
			Name:        "search",
			Description: "Search for information",
			Schema:      toolSchema,
		},
	}

	req := &runtime.Request{
		Instructions:   "Follow these system rules.",
		PromptTemplate: "",
		Input:          inputData,
		OutputSchema:   outputSchema,
		ToolSpecs:      tools,
	}

	builder := &runtime.PromptBuilder{}

	// Act
	prompt := builder.Build("What is AI?", req)

	// Assert
	if !strings.Contains(prompt, "SYSTEM INSTRUCTIONS") {
		t.Errorf("Expected SYSTEM INSTRUCTIONS section, got: %s", prompt)
	}

	if !strings.Contains(prompt, "[USER PROMPT]\n\nWhat is AI?") {
		t.Errorf("Expected user prompt in prompt text")
	}

	inputJSON, _ := json.Marshal(inputData)
	if !strings.Contains(prompt, string(inputJSON)) {
		t.Errorf("Expected input data in prompt")
	}

	if !strings.Contains(prompt, "[TOOLS]") {
		t.Errorf("Expected TOOLS section in prompt")
	}

	if !strings.Contains(prompt, `Tool: search`) {
		t.Errorf("Expected tool name 'search' in prompt")
	}

	if !strings.Contains(prompt, `"done"`) || !strings.Contains(prompt, `"out"`) {
		t.Errorf("Expected output schema properties in prompt")
	}

	if !strings.Contains(prompt, "[GUIDELINES]") {
		t.Errorf("Expected GUIDELINES section in prompt")
	}
}

func TestPromptBuilder_Build_SkipInput(t *testing.T) {
	req := &runtime.Request{
		Instructions: "Test skip input",
		SkipInput:    true,
		Input:        map[string]string{"ignored": "data"},
		OutputSchema: gojsonschema.NewStringLoader(`{"type": "object"}`),
	}

	builder := &runtime.PromptBuilder{}
	prompt := builder.Build("Check input skipping", req)

	if strings.Contains(prompt, "[INPUT]") {
		t.Errorf("Expected no INPUT section when SkipInput is true")
	}
}

func TestPromptBuilder_Build_NoTools(t *testing.T) {
	req := &runtime.Request{
		Instructions: "No tools available",
		Input:        map[string]string{"test": "value"},
		OutputSchema: gojsonschema.NewStringLoader(`{"type": "object"}`),
	}

	builder := &runtime.PromptBuilder{}
	prompt := builder.Build("Simple test", req)

	if strings.Contains(prompt, "[TOOLS]") {
		t.Errorf("Expected no TOOLS section when ToolSpecs is empty")
	}
}
