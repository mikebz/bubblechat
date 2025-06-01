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
	"os/exec"
	"strings"
	"testing"
)

func TestExecuteGcloudCommand(t *testing.T) {
	tests := []struct {
		name    string
		command string
		wantErr bool
		wantOut string
	}{
		{
			name:    "simple command",
			command: "gcloud version",
			wantErr: false,
			// We can't know the exact version, so we check for a common substring.
			// This makes the test less brittle.
			wantOut: "Google Cloud SDK",
		},
		{
			name:    "command with prefix",
			command: "gcloud version",
			wantErr: false,
			wantOut: "Google Cloud SDK",
		},
		{
			name:    "invalid command",
			command: "gcloud invalid-command-that-does-not-exist",
			wantErr: true,
			wantOut: "", // Error message will vary, so we don't check exact output
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip tests that require gcloud to be installed if it's not available.
			// This is a common pattern for tests that depend on external tools.
			if !isCommandAvailable("gcloud") && tt.name != "invalid command" {
				t.Skip("gcloud command not found, skipping test")
			}

			output, err := ExecuteGcloudCommand(tt.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteGcloudCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !strings.Contains(output, tt.wantOut) {
				t.Errorf("ExecuteGcloudCommand() output = %v, wantOut substring %v", output, tt.wantOut)
			}
			if tt.wantErr && output == "" && tt.name == "invalid command" {
                 // For invalid commands, we expect an error and potentially an empty output string
                 // or an output string containing the error message.
                 // If gcloud is not installed, 'output' might be empty and err will be set.
                 // If gcloud is installed, 'output' will contain the error from gcloud.
                 // The main check is that err is not nil (wantErr is true).
                 // We also check that output is not unexpectedly empty if an error IS expected.
                 // However, if gcloud is not installed, output will be empty.
                 // So, we only fail if output is empty AND gcloud is installed.
                 if isCommandAvailable("gcloud") && output == "" {
                    t.Errorf("ExecuteGcloudCommand() output was empty for an expected error, this might indicate an issue if gcloud is installed.")
                 }
            }
		})
	}
}

// isCommandAvailable checks if a command is available in the system PATH.
// This is a helper function for the tests.
func isCommandAvailable(name string) bool {
	// This is a simplified check. A more robust check might involve
	// searching through all directories in the PATH environment variable.
	// For now, we assume that if `gcloud help` runs without error, gcloud is available.
	// This is not perfect, as `gcloud help` might not be a valid command for all gcloud versions.
	// A better check would be `exec.LookPath("gcloud")`.
	// However, for the purpose of this generated test, we'll use a simpler approach.
	// We will try to run "gcloud help" and see if it errors.
	// This is not ideal because `gcloud help` itself is a gcloud command.
	// A truly robust check would be `exec.LookPath`.
	// Let's refine this to use `exec.LookPath`.
	_, err := exec.LookPath(name)
	return err == nil
}
