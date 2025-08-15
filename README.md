<p align="center">
<img alt="Logo" src="assets/logo.png" width="300px">
</p>
<h2 align="center">Type-Safe AI Agents for Go.</h2>

<p align="center">
  ⚠️ <strong>Note:</strong> Suricata is in early development. Some bugs may exist. Please report issues!


## Why Suricata?

Most LLM integrations rely on unstructured text—hard to maintain, easy to break, and impossible to type-check. **Suricata fixes this** by:

- Generating strongly typed Go code for every message and tool

- Providing compile-time guarantees with fewer runtime surprises

- Separating business logic from orchestration for cleaner code

Instead of wiring prompts and parsing JSON, you declare everything in YAML, generate Go stubs, and let Suricata handle orchestration.
Agents can dynamically choose tools at runtime—while you keep full control and type safety.

Adding a new tool? Define it in YAML and regenerate—no runtime edits needed. Suricata follows Go idioms, so your agents feel native.

**In short**: Suricata blends LLM intelligence, Go type safety, and a declarative workflow—giving you confidence to build production-ready AI agents.

## Features

- **Type-Safe by Design** – Define messages in YAML, generate Go types with compile-time guarantees.

- **Declarative Agents** – Describe behavior and prompts in YAML; Suricata handles orchestration.

- **Dynamic Tooling** – Register tools once, let agents choose at runtime.

- **Idiomatic Go Code** – Automatic stub generation, Go templates for dynamic prompts, and easy testing.

## Quickstart

Write an `hello-spec.yml` file:

```yaml

version: 0.0.1
package: example.hello

messages:
  SayHelloAllRequest:
    fields:
      - name: names
        type: string
        repeated: true
  SayHelloAllReply:
    fields:
      - name: ok
        type: bool
  SayHelloToolRequest:
    fields:
      - name: name
        type: string
  SayHelloToolReply:
    fields:
      - name: ok
        type: bool

tools:
  SayHelloTool:
    description: "Say hello to a given name"
    input: SayHelloToolRequest
    output: SayHelloToolReply

agents:
  HelloAgent:
    instructions: |
      You are a helpful and precise assistant. Your role is to say hello to people.
    actions:
      SayHelloAll:
        description: "Say hello to all names given as input"
        input: SayHelloAllRequest
        output: SayHelloAllReply
        prompt: |
           {{- /* Use Go templating for dynamic prompts */ -}}
          Please say hello to all the following names:
          {{- range .Names }}
          - {{ . }}
          {{- end }}
    tools:
      - SayHelloTool
```

Then, to generate the Go stubs, run:

```bash
suricata gen hello-spec.yml
```

Then, use your generated stubs:

```golang
package main

import (
  ...
  
	"github.com/ostafen/suricata/example/hello"
)

func main() {
	invoker := anthropic.NewAnthropicInvoker(APIKey, anthropic.ClaudeSonnet37, 1024)

	cli := v1.NewHelloAgent(invoker, &tools{})

	res, err := cli.SayHelloAll(context.Background(), &v1.SayHelloAllRequest{
		Names: []string{"Pippo", "Pluto"},
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(res.Ok)
}

type tools struct{}

func (t *tools) SayHelloTool(in *v1.SayHelloToolRequest) (*v1.SayHelloToolReply, error) {
	fmt.Println("Hello " + in.Name)

	return &v1.SayHelloToolReply{Ok: true}, nil
}
```

## 📄 License

`MIT` License. See `LICENSE` for details.