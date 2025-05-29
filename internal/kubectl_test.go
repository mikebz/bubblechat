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
