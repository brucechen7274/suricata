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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

type PromptBuilder struct {
	strings.Builder
}

func (pb *PromptBuilder) Build(userPrompt string, req *Request) string {
	pb.writeInstructions(req)

	if len(req.ToolSpecs) > 0 {
		pb.writeWorkflow()
	}

	pb.writeTools(req.ToolSpecs)

	if !req.SkipInput {
		pb.writeInput(req.Input)
	}

	pb.writeOutputFormat(req.OutputSchema, len(req.ToolSpecs) > 0)
	pb.writeGuidelines()
	pb.writeUserPrompt(userPrompt)

	return pb.String()
}

func (pb *PromptBuilder) writeInstructions(req *Request) {
	// System instructions
	if req.Instructions != "" {
		pb.WriteString("[SYSTEM INSTRUCTIONS]\n\n")
		pb.WriteString(req.Instructions)
		pb.WriteString("\n\n")
	}
}

func (pb *PromptBuilder) writeUserPrompt(prompt string) {
	// User prompt
	pb.WriteString("[USER PROMPT]\n\n")
	pb.WriteString(prompt)
	pb.WriteString("\n")
}

func (pb *PromptBuilder) writeWorkflow() {
	pb.WriteString(`
[WORKFLOW]

1. You will be given the conversation so far, including:
   - The original user request.
   - Your previous reasoning and tool calls.
   - Tool outputs or error messages.

2. After receiving a tool output or error, you must:
   - Analyze if the goal is achieved.
   - If more steps are required, call another tool with correct parameters.
   - If the goal is complete, provide a clear, final answer to the user.
`)
}

func (pb *PromptBuilder) writeInput(in any) {
	rawInput, _ := json.Marshal(in)
	pb.WriteString("\n[INPUT]:\n\n")
	pb.Write(rawInput)
	pb.WriteString("\n")
}

func (pb *PromptBuilder) writeTools(tools []ToolSpec) {
	if len(tools) > 0 {
		pb.WriteString("\n[TOOLS]\n\n")
		for _, tool := range tools {
			inSchema, _ := tool.Schema.LoadJSON()
			rawInSchema, _ := json.Marshal(inSchema)
			fmt.Fprintf(&pb.Builder, "Tool: %s\nDescription: %s\nInputSchema: %s\n\n", tool.Name, tool.Description, rawInSchema)
		}
	}
}

func (pb *PromptBuilder) writeOutputFormat(outSchema gojsonschema.JSONLoader, hasTools bool) {
	jsonSchema, _ := outSchema.LoadJSON()
	rawSchema, _ := json.Marshal(jsonSchema)

	if !hasTools {
		pb.WriteString(`
[OUTPUT FORMAT]

Return ONLY a valid JSON object matching the following schema:

` + string(rawSchema))
		return
	}

	pb.WriteString(`
[OUTPUT FORMAT]

After each tool output or error, you must return exactly one JSON object, following these rules:

1. If more steps are required (tool call):

{
	"name": "<tool name>",
	"args": {...}
}

- "name": The exact name of the tool to call (must be one of the tools listed in the TOOLS section).
- "args": A JSON object that matches the input schema for the selected tool exactly.
- Do not include extra fields or omit required fields.

2. If goal is achieved (final output):

{
	"done": true,
	"out": {...}
}

where "out" is a JSON object strictly matching the following JSON schema:

` + string(rawSchema))
}

func (pb *PromptBuilder) writeGuidelines() {
	pb.WriteString(`

[GUIDELINES]:

- Do not include any extra text.
- Do not include markdown or code fences.
- Ensure the JSON is syntactically valid.
- All fields must be present, even if empty.

`)
}
