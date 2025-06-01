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
)

// ExecuteGcloudCommand executes a gcloud command and returns the output or an error.
func ExecuteGcloudCommand(command string) (string, error) {
	// remove gcloud prefix if it exists
	command = strings.TrimPrefix(command, "gcloud ")
	cmd := exec.Command("gcloud", strings.Fields(command)...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
