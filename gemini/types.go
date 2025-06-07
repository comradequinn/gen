package gemini

type (
	Prompt struct {
		History            []Transaction
		InputType          InputType
		Text               string
		Files              []string
		CommandResult      CommandResult
		FilesRequestResult FilesRequestResult
		Schema             JSONSchema
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
		CommandResult      CommandResult      `json:"commandResult,omitzero"`
		FilesRequestResult FilesRequestResult `json:"filesRequestResponse,omitzero"`
	}
	Output struct {
		Text           string         `json:"text,omitempty,omitzero"`
		CommandRequest CommandRequest `json:"commandRequest,omitempty,omitzero"`
		FilesRequest   FilesRequest   `json:"filesRequest,omitempty,omitzero"`
	}

	Platform int
)

func (o Output) IsFunction() bool {
	return o.IsCommandRequest() || o.IsFilesRequest()
}

func (o Output) IsCommandRequest() bool {
	return o.CommandRequest.Text != ""
}

func (o Output) IsFilesRequest() bool {
	return len(o.FilesRequest.Files) > 0
}

func (i Input) IsCommandResult() bool {
	return i.CommandResult.Executed
}

func (i Input) IsFilesRequestResult() bool {
	return i.FilesRequestResult.Attached
}
