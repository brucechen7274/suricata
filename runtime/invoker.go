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
		system:   "", // TODO
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
