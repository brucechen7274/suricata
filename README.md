<p align="center">
<img alt="Logo" src="assets/logo.png" width="300px">
</p>
<h2 align="center">Suricata: Another agentic AI framework for Go - with type-safety this time.</h2>

<p align="center">
  ⚠️ <strong>Note:</strong> Suricata is in early development. Some bugs may exist. Please report issues!


## Why Suricata?

**Suricata** was built to make developing AI agents in Go safe, reliable, and maintainable. Unlike traditional LLM integrations that rely on unstructured data, Suricata generates Go types for every message and tool, giving you compile-time guarantees and reducing runtime errors.

Agents, actions, and prompts are defined declaratively in YAML, keeping your business logic separate from orchestration and making updates, testing, and versioning effortless. At runtime, agents can intelligently select from multiple tools, allowing LLMs to choose the right action while you retain full type safety and control.

Adding new tools or actions is simple: define them in YAML and regenerate stubs, enabling complex workflows without touching the runtime. Suricata follows Go conventions and idioms, so your agents feel like native Go code. Every tool is a Go interface, making testing and debugging straightforward.

In short, Suricata combines LLM intelligence, type safety, and Go-native ergonomics, giving developers the confidence to build robust, production-ready AI agents.

## 🌟 Features

- **Type-safe messages**: Define request and response messages and get generated Go types.

- **Declarative agents**: Describe agent behavior, actions, and prompts in a YAML file.

- **Tools integration**: Register multiple tools and let the LLM dynamically choose which one to call.

- **Code generation**: Generate idiomatic Go client stubs automatically.

- **Flexible LLM backend**: Compatible with different LLM invokers (Anthropic, OpenAI, etc.).

- **Easy testing & mocking**: Tools are interfaces, making unit tests simple.

## Quickstart

Write an `hello-spec.yml` file:

```yaml

version: llm-1
package: example.v1

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
          Say hello to all names given as input.
    tools:
      - SayHelloTool
```

Then, to generate the Go stubs, run:

```bash
suricata gen hello-spec.yml
```

Then, use your generated stubs:

```golang
func main() {
	invoker := anthropic.NewAnthropicInvoker(APIKey, anthropic.ClaudeSonnet37, 1024)

	cli := v1.NewHelloAgentClient(invoker, &tools{})

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