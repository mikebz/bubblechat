package internal

import (
	"os"
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

	// t.Context() is used for gollm.NewClient, but we don't need to import "context" for that.
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

// TestChatLoop does the basic test with a simple query
// expecting a text response without function calls.
func TestChatLoop(t *testing.T) {
	chat := setup(t, "", "")
	h := &History{
		Chat:    chat,
		Context: t.Context(), // h.Context is set using t.Context()
		Blocks:  []Block{},
	}

	// A query designed to elicit a simple text response without function calls.
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

// TestErrorChatLoop verifies that the ChatLoop method correctly handles errors
// by producing an ErrorBlock. It sets up a History with a potentially
// problematic chat configuration (e.g., an invalid model name "gemini-2.0-foobar")
// and an empty query. It then calls ChatLoop and asserts that the first
// block generated is of type ErrorBlock, indicating that the error was
// captured and recorded appropriately.
func TestErrorChatLoop(t *testing.T) {
	chat := setup(t, "gemini", "gemini-2.0-foobar")
	h := &History{
		Chat:    chat,
		Context: t.Context(), // h.Context is set using t.Context()
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
