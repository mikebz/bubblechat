package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"

	"github.com/GoogleCloudPlatform/kubectl-ai/gollm"
	in "github.com/mikebz/bubblechat/internal"
)

func main() {
	// Start the chat session
	fmt.Println("Welcome to BubbleChat! Type your message below:")

	ctx := context.Background()
	client, err := gollm.NewClient(ctx, "gemini")
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		return
	}

	err = in.Repl(ctx, client)
	if err != nil {
		os.Exit(1)

	}
}
