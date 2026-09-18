package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
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
	action := scenario.Match(prompt)
	if err := applyAction(action); err != nil {
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
