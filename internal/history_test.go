package internal

import (
	"context"
	// "fmt" // Removed: imported and not used
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/GoogleCloudPlatform/kubectl-ai/gollm"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

// setup that initializes a gollm.Chat client for integration tests.
func setup(t *testing.T, provider, model string) gollm.Chat {
	t.Helper()
	if provider == "" {
		provider = "gemini"
	}
	if model == "" {
		model = "gemini-2.0-flash"
	}
	err := godotenv.Load("../.env")
	if err != nil {
		if !os.IsNotExist(err) {
			t.Logf("Warning: Error loading .env file from ../.env: %v. Proceeding with environment variables.", err)
		} else {
			t.Logf("Info: .env file not found at ../.env. Relying on environment variables.")
		}
	}

	client, err := gollm.NewClient(t.Context(), provider)
	if err != nil {
		t.Fatalf("Failed to create LLM client: %v.", err)
	}

	llmChat := gollm.NewRetryChat(
		client.StartChat(systemPrompt, model),
		gollm.RetryConfig{
			MaxAttempts:    3,
			InitialBackoff: 10 * time.Second,
			MaxBackoff:     60 * time.Second,
			BackoffFactor:  2,
			Jitter:         true,
		},
	)
	return llmChat
}

func TestChatLoop(t *testing.T) {
	chat := setup(t, "", "")
	h := &History{
		Chat:    chat,
		Context: t.Context(),
		Blocks:  []Block{},
	}
	query := "Hello, this is a test query. Please provide a short text response without using any tools or functions."
	h.ChatLoop(query)
	if len(h.Blocks) == 0 {
		t.Fatalf("Expected at least one block after ChatLoop, got 0.")
	}
	t.Logf("ChatLoop resulted in %d block(s):", len(h.Blocks))
	for i, b := range h.Blocks {
		t.Logf("Block %d: %q", i, b.String())
	}
	firstBlock := h.Blocks[0]
	assert.Equal(t, AgentBlock, firstBlock.Type, "Expected first block to be a AgentBlock, got %s", firstBlock.Type)
}

func TestErrorChatLoop(t *testing.T) {
	chat := setup(t, "gemini", "gemini-2.0-foobar")
	h := &History{
		Chat:    chat,
		Context: t.Context(),
		Blocks:  []Block{},
	}
	query := ""
	h.ChatLoop(query)
	if len(h.Blocks) == 0 {
		t.Fatalf("Expected at least one block after ChatLoop, got 0.")
	}
	t.Logf("ChatLoop resulted in %d block(s):", len(h.Blocks))
	for i, b := range h.Blocks {
		t.Logf("Block %d: %q", i, b.String())
	}
	firstBlock := h.Blocks[0]
	assert.Equal(t, ErrorBlock, firstBlock.Type, "Expected first block to be an ErrorBlock, got %s", firstBlock.Type)
}

// --- Mock gollm types for testing function calls ---

// MockPart implements gollm.Part
type MockPart struct {
	isText       bool
	textContent  string
	isFuncCall   bool
	funcCalls    []gollm.FunctionCall
}

func (p *MockPart) AsText() (string, bool) {
	return p.textContent, p.isText
}

func (p *MockPart) AsFunctionCalls() ([]gollm.FunctionCall, bool) {
	return p.funcCalls, p.isFuncCall
}

// MockCandidate implements gollm.Candidate
type MockCandidate struct {
	parts []gollm.Part
}

func (c *MockCandidate) String() string      { return "mock candidate" }
func (c *MockCandidate) Parts() []gollm.Part { return c.parts }

// MockChatResponse implements gollm.ChatResponse
type MockChatResponse struct {
	candidates []gollm.Candidate
}

func (r *MockChatResponse) UsageMetadata() any             { return nil }
func (r *MockChatResponse) Candidates() []gollm.Candidate { return r.candidates }

// newMockChatResponseWithFunctionCall creates a gollm.ChatResponse with a single function call.
func newMockChatResponseWithFunctionCall(fc gollm.FunctionCall) gollm.ChatResponse {
	return &MockChatResponse{
		candidates: []gollm.Candidate{
			&MockCandidate{
				parts: []gollm.Part{
					&MockPart{
						isFuncCall: true,
						funcCalls:  []gollm.FunctionCall{fc},
					},
				},
			},
		},
	}
}

// newMockChatResponseWithText creates a gollm.ChatResponse with simple text.
func newMockChatResponseWithText(text string) gollm.ChatResponse {
	return &MockChatResponse{
		candidates: []gollm.Candidate{
			&MockCandidate{
				parts: []gollm.Part{
					&MockPart{
						isText:      true,
						textContent: text,
					},
				},
			},
		},
	}
}

// MockChat is a mock implementation of gollm.Chat for testing.
type MockChat struct {
	NextResponse gollm.ChatResponse // The response to return on the next call to Send
	SentMessages []any              // Records messages sent via Send (contents ...any)
}

// Send implements the gollm.Chat interface.
func (mc *MockChat) Send(ctx context.Context, contents ...any) (gollm.ChatResponse, error) {
	if len(contents) > 0 {
		mc.SentMessages = append(mc.SentMessages, contents[0])
	} else {
		mc.SentMessages = append(mc.SentMessages, "")
	}

	if mc.NextResponse == nil {
		return newMockChatResponseWithText("mocked default text response"), nil
	}
	return mc.NextResponse, nil
}

// SendStreaming implements the gollm.Chat interface.
func (mc *MockChat) SendStreaming(ctx context.Context, contents ...any) (gollm.ChatResponseIterator, error) {
	// This mock doesn't currently support streaming. Return nil for the iterator.
	// This will satisfy the interface, but panic if called and used.
	// The tests being added do not call SendStreaming.
	return nil, nil
}

// Start implements the gollm.Chat interface.
// Note: The gollm.Chat interface itself doesn't list Start, but it's implied by Client.StartChat returning Chat.
// For a mock, it's good to have it if other parts of the code might expect to call Start on a Chat instance.
func (mc *MockChat) Start(ctx context.Context, systemMessage string, options ...any) error {
	return nil
}

// SetFunctionDefinitions implements the gollm.Chat interface.
func (mc *MockChat) SetFunctionDefinitions(functionDefinitions []*gollm.FunctionDefinition) error {
	return nil
}

// IsRetryableError implements the gollm.Chat interface.
func (mc *MockChat) IsRetryableError(err error) bool {
	return false
}

var execLookPathForHistoryTest = exec.LookPath

// isToolCommandAvailableInHistoryTest checks if a command is available for history tests.
func isToolCommandAvailableInHistoryTest(name string) bool {
	_, err := execLookPathForHistoryTest(name)
	return err == nil
}

func TestChatLoop_KubectlCommand_Success(t *testing.T) {
	if !isToolCommandAvailableInHistoryTest("kubectl") {
		t.Skip("kubectl command not found, skipping TestChatLoop_KubectlCommand_Success")
	}

	mockChat := &MockChat{}
	history := &History{
		Blocks:  []Block{},
		Chat:    mockChat, // This should now be valid
		Context: context.Background(),
	}

	mockChat.NextResponse = newMockChatResponseWithFunctionCall(
		gollm.FunctionCall{
			Name: "kubectl",
			Arguments: map[string]any{
				"command": "version --client",
			},
		},
	)

	history.ChatLoop("user query to trigger kubectl")

	var toolOutputBlock *Block
	var functionCallBlock *Block

	for i := range history.Blocks {
		if history.Blocks[i].Type == ToolBlock && strings.HasPrefix(history.Blocks[i].Text, "Function: kubectl") {
			functionCallBlock = &history.Blocks[i]
			if i+1 < len(history.Blocks) && history.Blocks[i+1].Type == ToolBlock {
				toolOutputBlock = &history.Blocks[i+1]
			}
			break
		}
	}

	if functionCallBlock == nil {
		t.Fatalf("Expected a 'Function: kubectl' block in history, got none. History: %v", history.Blocks)
	}
	if toolOutputBlock == nil {
		t.Fatalf("Expected a ToolBlock with kubectl output after function call block, got none or wrong type. History: %v", history.Blocks)
	}
	if !strings.Contains(toolOutputBlock.Text, "Client Version:") {
		t.Errorf("Expected kubectl version output to contain 'Client Version:', got: %s", toolOutputBlock.Text)
	}
}

func TestChatLoop_GcloudCommand_Success(t *testing.T) {
	if !isToolCommandAvailableInHistoryTest("gcloud") {
		t.Skip("gcloud command not found, skipping TestChatLoop_GcloudCommand_Success")
	}

	mockChat := &MockChat{}
	history := &History{
		Blocks:  []Block{},
		Chat:    mockChat, // This should now be valid
		Context: context.Background(),
	}

	mockChat.NextResponse = newMockChatResponseWithFunctionCall(
		gollm.FunctionCall{
			Name: "gcloud",
			Arguments: map[string]any{
				"command": "version",
			},
		},
	)

	history.ChatLoop("user query to trigger gcloud")

	var toolOutputBlock *Block
	var functionCallBlock *Block

	for i := range history.Blocks {
		if history.Blocks[i].Type == ToolBlock && strings.HasPrefix(history.Blocks[i].Text, "Function: gcloud") {
			functionCallBlock = &history.Blocks[i]
			if i+1 < len(history.Blocks) && history.Blocks[i+1].Type == ToolBlock {
				toolOutputBlock = &history.Blocks[i+1]
			}
			break
		}
	}

	if functionCallBlock == nil {
		t.Fatalf("Expected a 'Function: gcloud' block in history, got none. History: %v", history.Blocks)
	}
	if toolOutputBlock == nil {
		t.Fatalf("Expected a ToolBlock with gcloud output after function call block, got none or wrong type. History: %v", history.Blocks)
	}
	if !strings.Contains(toolOutputBlock.Text, "Google Cloud SDK") {
		t.Errorf("Expected gcloud version output to contain 'Google Cloud SDK', got: %s", toolOutputBlock.Text)
	}
}

func TestChatLoop_ToolCommand_Failure(t *testing.T) {
	if !isToolCommandAvailableInHistoryTest("kubectl") {
		t.Skip("kubectl command not found, skipping TestChatLoop_ToolCommand_Failure")
	}

	mockChat := &MockChat{}
	history := &History{
		Blocks:  []Block{},
		Chat:    mockChat, // This should now be valid
		Context: context.Background(),
	}

	mockChat.NextResponse = newMockChatResponseWithFunctionCall(
		gollm.FunctionCall{
			Name: "kubectl",
			Arguments: map[string]any{
				"command": "nonexistent-command arg1 arg2",
			},
		},
	)

	history.ChatLoop("user query to trigger failing kubectl command")

	var errorBlock *Block
	var functionCallBlock *Block

	for i := range history.Blocks {
		if history.Blocks[i].Type == ToolBlock && strings.HasPrefix(history.Blocks[i].Text, "Function: kubectl") {
			functionCallBlock = &history.Blocks[i]
			if i+1 < len(history.Blocks) && history.Blocks[i+1].Type == ErrorBlock {
				errorBlock = &history.Blocks[i+1]
			}
			break
		}
	}

	if functionCallBlock == nil {
		t.Fatalf("Expected a 'Function: kubectl' block in history, got none. History: %v", history.Blocks)
	}
	if errorBlock == nil {
		t.Fatalf("Expected an ErrorBlock after a failing kubectl command, got none or wrong type. History: %v", history.Blocks)
	}
	if !strings.Contains(errorBlock.Text, "Error executing kubectl") {
		t.Errorf("Expected error block to contain 'Error executing kubectl', got: %s", errorBlock.Text)
	}
	if !strings.Contains(errorBlock.Text, "Output:") {
		t.Errorf("Expected error block to contain 'Output:', got: %s", errorBlock.Text)
	}
}

func TestChatLoop_UnknownTool(t *testing.T) {
	mockChat := &MockChat{}
	history := &History{
		Blocks:  []Block{},
		Chat:    mockChat, // This should now be valid
		Context: context.Background(),
	}

	mockChat.NextResponse = newMockChatResponseWithFunctionCall(
		gollm.FunctionCall{
			Name: "unknown-tool",
			Arguments: map[string]any{
				"command": "some arguments",
			},
		},
	)

	history.ChatLoop("user query to trigger unknown tool")

	var errorBlock *Block
	var functionCallBlock *Block

	for i := range history.Blocks {
		if history.Blocks[i].Type == ToolBlock && strings.HasPrefix(history.Blocks[i].Text, "Function: unknown-tool") {
			functionCallBlock = &history.Blocks[i]
			if i+1 < len(history.Blocks) && history.Blocks[i+1].Type == ErrorBlock {
				errorBlock = &history.Blocks[i+1]
			}
			break
		}
	}

	if functionCallBlock == nil {
		t.Fatalf("Expected a 'Function: unknown-tool' block in history, got none. History: %v", history.Blocks)
	}
	if errorBlock == nil {
		t.Fatalf("Expected an ErrorBlock after an unknown tool call, got none or wrong type. History: %v", history.Blocks)
	}
	if errorBlock.Type != ErrorBlock {
		t.Errorf("Expected block type ErrorBlock for unknown tool, got %v. History: %v", errorBlock.Type, history.Blocks)
	}
	if !strings.Contains(errorBlock.Text, "Error executing unknown-tool") {
		t.Errorf("Expected error block to contain 'Error executing unknown-tool', got: %s", errorBlock.Text)
	}
	if !strings.Contains(errorBlock.Text, "Output: Unknown tool: unknown-tool") {
		t.Errorf("Expected error block to contain 'Output: Unknown tool: unknown-tool', got: %s", errorBlock.Text)
	}
}
