package internal

import (
	"context"

	"github.com/GoogleCloudPlatform/kubectl-ai/gollm"
)

// BlockType defines the type of a block in the conversation.
type BlockType int

const (
	// ErrorBlock indicates an error message.
	ErrorBlock BlockType = iota
	// AgentBlock indicates a message from the AI agent.
	AgentBlock
	// UserBlock indicates a message from the user.
	UserBlock
	// ToolBlock indicates a message from a tool.
	ToolBlock
)

// Block represents a single message block in the conversation.
// It can be a user message, an AI response, an error message, or a tool response.
type Block struct {
	Text string
	Type BlockType
}

// History represents the conversation history, which is a collection of blocks.
type History struct {
	Blocks  []Block
	Chat    gollm.Chat
	Context context.Context
}

func (h *History) AddBlock(block Block) {
	h.Blocks = append(h.Blocks, block)
}
