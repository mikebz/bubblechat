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
	_ "embed"
	"fmt"
	"strings"

	"github.com/GoogleCloudPlatform/kubectl-ai/gollm"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

//go:embed systemprompt.txt
var systemPrompt string

// Repl starts the REPL (Read-Eval-Print Loop) for the BubbleChat application.
// It initializes the chat client, sets up the conversation history,
// and starts the BubbleTea program to handle user input and display the conversation.
// It's expected that outside of initializing the client you will not need to do
// anything to have an interactive session.
func Repl(ctx context.Context, client gollm.Client) error {

	doc := NewDoc(ctx, client)

	p := tea.NewProgram(doc)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		return err
	}

	return nil
}

var (
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#cc0000"))
	agentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#d3d7cf"))
	userStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#729fcf"))
	toolStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#32afff"))
	otherStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ad7fa8"))
)

// Render formats a Block for display in the terminal.
// It applies different styles based on the type of block (e.g., error, agent, user, tool).
func Render(block Block) string {

	var lgStyle lipgloss.Style
	switch block.Type {
	case ErrorBlock:
		lgStyle = errorStyle
	case AgentBlock:
		out, err := glamour.Render(block.Text, "dark")
		if err == nil {
			return out
		}
		lgStyle = agentStyle
	case UserBlock:
		lgStyle = userStyle
	case ToolBlock:
		lgStyle = toolStyle
	default:
		lgStyle = otherStyle
	}

	return lgStyle.Render(block.Text)
}

// Document also contains visual elements in addition
// to the conversation history datamodel.
type Document struct {
	History
	textInput textinput.Model
}

func NewDoc(context context.Context, client gollm.Client) *Document {
	history := NewHistory(context, client, "")
	doc := &Document{
		textInput: textinput.New(),
	}
	doc.textInput.Focus()
	doc.History = *history
	doc.AddBlock(Block{
		Text: "Welcome to BubbleChat! Type your message below:",
		Type: AgentBlock,
	})

	return doc
}

// HandleSend processes the user input when the Enter key is pressed.
// It adds the user input as a new block in the conversation history
// and sends the message to the chat service.
func (doc *Document) HandleSend() {
	userInput := strings.TrimSpace(doc.textInput.Value())
	if userInput == "" {
		return
	}

	doc.AddBlock(Block{
		Text: userInput,
		Type: UserBlock,
	})
	doc.textInput.Reset()
	doc.textInput.Focus()

	doc.ChatLoop(userInput)
}

// Init initializes the text input model and returns a command to start blinking the cursor.
// This function is called by BubbleTea when the program starts.
func (doc *Document) Init() tea.Cmd {
	return textinput.Blink
}

// View renders the current state of the document, including the conversation history.
// This function is called by BubbleTea to display the UI.
func (doc *Document) View() string {
	var sb strings.Builder
	for _, block := range doc.Blocks {
		sb.WriteString(Render(block))
		sb.WriteString("\n")
	}
	sb.WriteString(doc.textInput.View())
	sb.WriteString("\nPress Ctrl+C or Esc to exit.\n")
	return sb.String()
}

// Update processes incoming messages and user input.
// This function is called by BubbleTea whenever there is a new message or user input.
func (doc *Document) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return doc, tea.Quit
		case tea.KeyEnter:
			doc.HandleSend()
			return doc, nil
		}
	}

	var cmd tea.Cmd
	doc.textInput, cmd = doc.textInput.Update(msg)

	return doc, cmd
}
