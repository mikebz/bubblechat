package internal

import (
	"os"
	"testing"

	"github.com/GoogleCloudPlatform/kubectl-ai/gollm"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

// setup that initializes a gollm.Chat client for integration tests.
func setup(t *testing.T, provider, model string) *History {
	t.Helper()
	if provider == "" {
		provider = "gemini"
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

	h := NewHistory(t.Context(), client, model)
	return h
}

// TestChatLoop does the basic test with a simple query
// expecting a text response without function calls.
func TestChatLoop(t *testing.T) {
	h := setup(t, "", "")

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
	h := setup(t, "gemini", "gemini-2.0-foobar")
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

// TestKubectlCall tests the ChatLoop method with a query that is expected
// to trigger a kubectl command execution. It sets up a History with a
// valid chat configuration and a query that requests the namespaces in the cluster.
func TestKubectlCall(t *testing.T) {
	h := setup(t, "", "")

	// A query designed to elicit a simple text response without function calls.
	query := "Please get the namespaces on my cluster"
	h.ChatLoop(query)
	if len(h.Blocks) == 0 {
		t.Fatalf("Expected at least one block after ChatLoop, got 0.")
	}
	t.Logf("ChatLoop resulted in %d block(s):", len(h.Blocks))

	// get the tool block
	var toolBlock *Block
	for i, b := range h.Blocks {
		t.Logf("Block %d: %q", i, b.String())
		// get the first tool block

		if toolBlock == nil && b.Type == ToolBlock {
			toolBlock = &b
		}
	}
	assert.Contains(t, toolBlock.Text, "kubectl", "Expected tool block to contain 'kubectl get namespaces', got %s", toolBlock.Text)
}
