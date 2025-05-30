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
	"container/list"
	"context"
	"fmt"

	"github.com/GoogleCloudPlatform/kubectl-ai/gollm"
)

// BlockType defines the type of a block in the conversation.
type BlockType int

const chatIterations = 5

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

func (b *Block) String() string {
	switch b.Type {
	case ErrorBlock:
		return fmt.Sprintf("Error: %s", b.Text)
	case AgentBlock:
		return fmt.Sprintf("AI: %s", b.Text)
	case UserBlock:
		return fmt.Sprintf("User: %s", b.Text)
	case ToolBlock:
		return fmt.Sprintf("Tool: %s", b.Text)
	default:
		return fmt.Sprintf("Unknown Block Type: %s", b.Text)
	}
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

func (h *History) ChatLoop(query string) {

	// Add the user's query to the conversation history
	resp, err := h.Chat.Send(h.Context, query)
	if err != nil {
		h.AddBlock(Block{
			Text: fmt.Sprintf("Error: %v", err),
			Type: ErrorBlock,
		})
		return
	}

	for i := 0; i < chatIterations; i++ {
		// If the response is empty, break the loop
		if resp == nil {
			return
		}

		if len(resp.Candidates()) == 0 {
			h.AddBlock(Block{
				Text: "No response from the AI agent.",
				Type: ErrorBlock,
			})
			return
		}

		candidate := resp.Candidates()[0]
		if len(candidate.Parts()) == 0 {
			h.AddBlock(Block{
				Text: "No response parts from the AI agent.",
				Type: ErrorBlock,
			})
			return
		}

		// Reset response for the next iteration
		resp = nil

		queue := list.New()

		// this loop goes through the parts.
		// insfoar as they are function calls we will add them to the
		// calls to process later.  insofar as they are text, we will
		// add them to the history.
		for _, part := range candidate.Parts() {
			fncalls, success := part.AsFunctionCalls()
			if success {
				for _, fncall := range fncalls {
					h.AddBlock(Block{
						Text: fmt.Sprintf("Function: %s", fncall.Name),
						Type: ToolBlock,
					})
					queue.PushBack(fncall)
				}
			} else {
				text, success := part.AsText()
				if success {
					h.AddBlock(Block{
						Text: text,
						Type: AgentBlock,
					})
				} else {
					h.AddBlock(Block{
						Text: "Unknown part type in response.",
						Type: ErrorBlock,
					})
				}

			}
		}

		// Process the function calls in the queue
		for queue.Len() > 0 {
			element := queue.Front()
			queue.Remove(element)
			fnCall := element.Value.(gollm.FunctionCall)

			// TODO: make the actual call
			// assign the response to resp
			h.AddBlock(Block{
				Text: fmt.Sprintf("Please call: %s, %v", fnCall.Name, fnCall.Arguments),
				Type: ToolBlock,
			})
		}

	}

}
