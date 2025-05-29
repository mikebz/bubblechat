package internal

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/kubectl-ai/gollm"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Document also contains visual elements in addition
// to the conversation history datamodel.
type Document struct {
	History
	textInput textinput.Model
}

func NewDoc(context context.Context, chat gollm.Chat) *Document {
	doc := &Document{
		History: History{
			Blocks:  []Block{},
			Chat:    chat,
			Context: context,
		},
		textInput: textinput.New(),
	}
	doc.textInput.Focus()

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

	resp, err := doc.Chat.Send(doc.Context, userInput)
	if err != nil {
		doc.AddBlock(Block{
			Text: fmt.Sprintf("Error: %v", err),
			Type: ErrorBlock,
		})
		return
	}

	agentResponse, _ := resp.Candidates()[0].Parts()[0].AsText()
	doc.AddBlock(Block{
		Text: agentResponse,
		Type: AgentBlock,
	})
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
		switch block.Type {
		case ErrorBlock:
			sb.WriteString(fmt.Sprintf("[Error] %s\n", block.Text))
		case AgentBlock:
			sb.WriteString(fmt.Sprintf("[Agent] %s\n", block.Text))
		case UserBlock:
			sb.WriteString(fmt.Sprintf("[User] %s\n", block.Text))
		case ToolBlock:
			sb.WriteString(fmt.Sprintf("[Tool] %s\n", block.Text))
		default:
			sb.WriteString(fmt.Sprintf("[Unknown] %s\n", block.Text))
		}
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

//go:embed systemprompt.txt
var systemPrompt string

// repl is a read-eval-print loop for the chat session.
func Repl(ctx context.Context, client gollm.Client) error {
	llmChat := gollm.NewRetryChat(
		client.StartChat(systemPrompt, "gemini-2.0-flash"),
		gollm.RetryConfig{
			MaxAttempts:    3,
			InitialBackoff: 10 * time.Second,
			MaxBackoff:     60 * time.Second,
			BackoffFactor:  2,
			Jitter:         true,
		},
	)

	doc := NewDoc(ctx, llmChat)

	p := tea.NewProgram(doc)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		return err
	}

	return nil
}
