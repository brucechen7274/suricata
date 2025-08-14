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
)

// Request holds the input parameters for an LLM call.
type Request struct {
	SkipInput      bool
	Instructions   string
	PromptTemplate string // Go template string for the prompt
	Input          any    // Data passed to the prompt template
	Output         any
	OutputSchema   gojsonschema.JSONLoader // Pointer to struct to unmarshal output JSON into

	ToolUnmarshaller ToolUnmarshaller
	ToolInvoker      ToolInvoker
	ToolSpecs        []ToolSpec
}

type Runtime struct {
	chat *ChatSession
}

func NewRuntime(chat *ChatSession) *Runtime {
	return &Runtime{chat: chat}
}

type ToolSpec struct {
	Name        string
	Description string
	Schema      gojsonschema.JSONLoader
}

type ToolResponse struct {
	Done bool `json:"done"`
	Out  any  `json:"out"`

	Name string `json:"name"`
	Args any    `json:"args"`
}

func (r *Runtime) Invoke(ctx context.Context, req Request) error {
	compiledPrompt, err := r.compilePrompt(&req)
	if err != nil {
		return err
	}

	prompt := getPrompt(compiledPrompt, &req)

	out, err := r.chat.Invoke(
		ctx,
		prompt,
	)
	if err != nil {
		return err
	}

	if req.ToolInvoker == nil {
		return unmarshalOutput(out, &req)
	}
	return r.agentLoop(ctx, out, &req)
}

func (r *Runtime) agentLoop(ctx context.Context, out string, req *Request) error {
	var resp ToolResponse

	for {
		out = ExtractJSONFromString(out)
		if err := json.Unmarshal([]byte(out), &resp); err != nil {
			return err
		}

		if resp.Done {
			break
		}

		rawArgs, _ := json.Marshal(resp.Args)
		inType, err := req.ToolUnmarshaller(resp.Name, rawArgs)
		if err != nil {
			return err
		}

		toolOutput := r.callTool(ctx, resp.Name, inType, req)

		out, err = r.chat.Invoke(ctx, toolOutput)
		if err != nil {
			return err
		}
	}

	rawOut, _ := json.Marshal(resp.Out)
	return unmarshalOutput(string(rawOut), req)
}

func (r *Runtime) callTool(ctx context.Context, name string, inType any, req *Request) string {
	toolResp, err := req.ToolInvoker(ctx, name, inType)
	if err != nil {
		return "ERR: " + err.Error()
	}

	rawToolResp, _ := json.Marshal(toolResp)
	return "OUTPUT: " + string(rawToolResp)
}

func unmarshalOutput(out string, req *Request) error {
	out = ExtractJSONFromString(out)
	if out == "" {
		return ErrInvalidOutput
	}

	res, err := gojsonschema.Validate(req.OutputSchema, gojsonschema.NewStringLoader(out))
	if err != nil {
		return err
	}

	if !res.Valid() {
		return ErrInvalidOutput
	}
	return json.Unmarshal([]byte(out), req.Output)
}

func getPrompt(userPrompt string, req *Request) string {
	prompt := getInstructions(req.Instructions)

	prompt += `
USER PROMPT:

` + userPrompt

	if req.ToolInvoker != nil {
		prompt += `
WORKFLOW:

1. You will be given the conversation so far, including:
   - The original user request.
   - Your previous reasoning and tool calls.
   - Tool outputs or error messages.

2. After receiving a tool output or error, you must:
   - Analyze if the goal is achieved.
   - If more steps are required, call another tool with correct parameters.
   - If the goal is complete, provide a clear, final answer to the user.
`
	}

	if !req.SkipInput {
		rawInput, _ := json.Marshal(req.Input)
		prompt += "\nINPUT:\n\n" + string(rawInput) + "\n"
	}

	outSchemaJSON, _ := req.OutputSchema.LoadJSON()
	rawSchema, _ := json.Marshal(outSchemaJSON)

	prompt += getToolsSection(req.ToolSpecs)
	prompt += getOutputSection(string(rawSchema), req.ToolInvoker != nil)

	prompt += `

GUIDELINES:

- Do not include any extra text.
- Do not include markdown or code fences.
- Ensure the JSON is syntactically valid.
- All fields must be present, even if empty.
`
	return prompt
}

func getInstructions(instructions string) string {
	if instructions == "" {
		return ""
	}
	return "SYSTEM INSTRUCTIONS:\n\n" + instructions + "\n\n"
}

func getOutputSection(outSchema string, hasTools bool) string {
	if !hasTools {
		return `
OUTPUT FORMAT:

Return ONLY a valid JSON object that matches the following JSON schema:

` + outSchema
	}

	return `
OUTPUT FORMAT:

After each tool output or error, you must return exactly one JSON object, following these rules:

1. If more steps are required (another tool call):

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

` + outSchema
}

func getToolsSection(tools []ToolSpec) string {
	if len(tools) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n[TOOLS]\n\n")

	for _, tool := range tools {
		inSchema, _ := tool.Schema.LoadJSON()
		rawInSchema, _ := json.Marshal(inSchema)

		fmt.Fprintf(&sb, "- Name: %s\n Description: %s\n InputSchema: %s\n\n", tool.Name, tool.Description, rawInSchema)
	}
	return sb.String()
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
