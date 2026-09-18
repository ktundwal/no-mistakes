package main

import (
	"path/filepath"
	"strconv"
	"testing"
)

func TestCopilotPromptWorkDir(t *testing.T) {
	target := filepath.Join(t.TempDir(), "target repo")
	prompt := "## no-mistakes isolated target\n\n" +
		"The repository you must inspect and modify is at " + strconv.Quote(target) + ". " +
		"Perform every file and shell operation against that exact directory."
	if got := copilotPromptWorkDir(prompt); got != target {
		t.Fatalf("copilotPromptWorkDir = %q, want %q", got, target)
	}
}
