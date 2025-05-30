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
	// Attempt to load .env file from project root (assuming tests run from package dir, e.g., internal/)
	err := godotenv.Load("../.env")
	if err != nil {
		if !os.IsNotExist(err) {
			// Log if error is not "file does not exist"
			t.Logf("Warning: Error loading .env file from ../.env: %v. Proceeding with environment variables.", err)
		} else {
			t.Logf("Info: .env file not found at ../.env. Relying on environment variables.")
		}
	}

	// TODO make these configurable via flags or env vars
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
		Context: t.Context(),
		Blocks:  []Block{},
	}

	// A query designed to elicit a simple text response without function calls.
	query := "Hello, this is a test query. Please provide a short text response without using any tools or functions."
	h.ChatLoop(query)

	if len(h.Blocks) == 0 {
		t.Fatalf("Expected at least one block after ChatLoop, got 0. This might indicate an issue with Chat.Send or ChatLoop's error handling for empty responses.")
	}

	t.Logf("ChatLoop resulted in %d block(s):", len(h.Blocks))
	for i, b := range h.Blocks {
		// Using %#v for Type to see the underlying integer value if needed for debugging.
		t.Logf("Block %d: %q", i, b.String())
	}

	firstBlock := h.Blocks[0]
	assert.Equal(t, firstBlock.Type, AgentBlock, "Expected first block to be a AgentBlock, got %s", firstBlock.Type)
}

func TestErrorChatLoop(t *testing.T) {
	// This test is designed to trigger an error response from the chat.
	chat := setup(t, "gemini", "gemini-2.0-foobar")

	h := &History{
		Chat:    chat,
		Context: t.Context(),
		Blocks:  []Block{},
	}

	// A query designed to elicit an error response.
	query := ""
	h.ChatLoop(query)

	if len(h.Blocks) == 0 {
		t.Fatalf("Expected at least one block after ChatLoop, got 0. This might indicate an issue with Chat.Send or ChatLoop's error handling for empty responses.")
	}

	t.Logf("ChatLoop resulted in %d block(s):", len(h.Blocks))
	for i, b := range h.Blocks {
		t.Logf("Block %d: %q", i, b.String())
	}
	firstBlock := h.Blocks[0]
	assert.Equal(t, firstBlock.Type, ErrorBlock, "Expected first block to be an ErrorBlock, got %s", firstBlock.Type)
}
