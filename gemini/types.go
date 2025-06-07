package gemini

import (
	"encoding/json"
	"fmt"
)

type (
	Config struct {
		GeminiURL      string
		Credential     string
		Platform       Platform
		FileStorageURL string
		SystemPrompt   string
		MaxTokens      int
		Temperature    float64
		TopP           float64
		UseCase        string
		Grounding      bool
		Agentic        bool
	}
	Prompt struct {
		History            []Transaction
		InputType          InputType
		Text               string
		Files              []string
		ShellCommandResult ShellCommandResult
		Schema             JSONSchema
	}
	ShellCommandResult struct {
		Code   int
		StdErr string
		StdOut string
	}
	FileReference struct {
		URI      string `json:"uri"`
		MIMEType string `json:"mimeType"`
		Label    string `json:"label"`
	}
	Transaction struct {
		Tokens int    `json:"tokens"`
		Input  Input  `json:"input"`
		Output Output `json:"output"`
	}
	Role       string
	InputType  string
	JSONSchema string
	Input      struct {
		Type               InputType          `json:"type"`
		Text               string             `json:"text,omitempty,omitzero"`
		FileReferences     []FileReference    `json:"files,omitempty,omitzero"`
		ShellCommandResult ShellCommandResult `json:"shellCommandOutput,omitempty,omitzero"`
	}
	Output struct {
		Text     string     `json:"text,omitempty,omitzero"`
		Function JSONSchema `json:"function,omitempty,omitzero"`
	}

	Platform int
)

func (c ShellCommandResult) marshalJSON() (json.RawMessage, error) {
	j, err := json.Marshal(map[string]any{
		"name": "local-shell-command",
		"response": map[string]any{
			"returnCode": c.Code,
			"stdErr":     c.StdErr,
			"stdOut":     c.StdOut,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("unable to marshal shell command output result into json. %w", err)
	}

	return j, nil
}
