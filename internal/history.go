// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
