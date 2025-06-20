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
