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
	"testing"
)

// Mock function for executing kubectl commands
func execCommand(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func TestKubectlGetPods(t *testing.T) {
	output, err := execCommand("kubectl", "get", "pods")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if output == "" {
		t.Fatal("Expected output, got empty string")
	}
}

func TestKubectlGetServices(t *testing.T) {
	output, err := execCommand("kubectl", "get", "services")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if output == "" {
		t.Fatal("Expected output, got empty string")
	}
}
