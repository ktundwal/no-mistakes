package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const copilotProjectInstructionProbe = "FAKEAGENT_PROJECT_INSTRUCTION_LOADED"

// runCopilot is a synthetic JSONL fixture for dispatch tests. When the target
// checkout contains the probe instruction, it obeys it unless the production
// suppression flag is present. This makes the e2e journey fail if no-mistakes
// ever launches Copilot with project instructions enabled under the trusted
// opt-out.
func runCopilot(args []string, input io.Reader, scenario *Scenario) int {
	data, err := io.ReadAll(input)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	prompt := string(data)
	logInvocation("copilot", prompt, args)
	workDir := copilotPromptWorkDir(prompt)
	action := scenario.MatchInDir(workDir, prompt)
	if err := applyActionInDir(workDir, action); err != nil {
		return 1
	}

	response := string(action.structuredJSON())
	if !hasExactArg(args, "--no-custom-instructions") {
		if instructions, readErr := os.ReadFile("AGENTS.md"); readErr == nil && strings.Contains(string(instructions), copilotProjectInstructionProbe) {
			response = copilotProjectInstructionProbe
		}
	}

	enc := json.NewEncoder(os.Stdout)
	_ = enc.Encode(map[string]any{
		"type": "assistant.message",
		"data": map[string]any{"content": response, "outputTokens": 50},
	})
	_ = enc.Encode(map[string]any{"type": "result", "exitCode": 0})
	return 0
}

func hasExactArg(args []string, want string) bool {
	for _, arg := range args {
		if arg == want {
			return true
		}
	}
	return false
}

func copilotHooksDisabled() bool {
	home := os.Getenv("COPILOT_HOME")
	if home == "" {
		return false
	}
	data, err := os.ReadFile(filepath.Join(home, "settings.json"))
	if err != nil {
		return false
	}
	var settings struct {
		DisableAllHooks bool `json:"disableAllHooks"`
	}
	return json.Unmarshal(data, &settings) == nil && settings.DisableAllHooks
}

func copilotPromptWorkDir(prompt string) string {
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}
	const (
		prefix = "The repository you must inspect and modify is at "
		suffix = ". Perform every file and shell operation"
	)
	start := strings.Index(prompt, prefix)
	if start < 0 {
		return wd
	}
	quoted := prompt[start+len(prefix):]
	end := strings.Index(quoted, suffix)
	if end < 0 {
		return wd
	}
	target, err := strconv.Unquote(quoted[:end])
	if err != nil || !filepath.IsAbs(target) {
		return wd
	}
	return target
}
