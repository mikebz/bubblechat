package internal

import (
	"os/exec"
	"strings"
)

// ExecuteKubectlCommand executes a kubectl command and returns the output or an error.
func ExecuteKubectlCommand(command string) (string, error) {
	// remove kubectl prefix if it exists
	command = strings.TrimPrefix(command, "kubectl ")
	cmd := exec.Command("kubectl", strings.Fields(command)...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
