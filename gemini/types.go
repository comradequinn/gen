package gemini

type (
	Prompt struct {
		History       []Transaction
		InputType     InputType
		Text          string
		FilePaths     []string
		ExecuteResult ExecuteResult
		ReadResult    ReadResult
		WriteResult   WriteResult
		Schema        JSONSchema
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
		Type               InputType       `json:"type"`
		Text               string          `json:"text,omitempty,omitzero"`
		FileReferences     []FileReference `json:"files,omitempty,omitzero"`
		ExecuteResult      ExecuteResult   `json:"executeResult,omitzero"`
		FilesRequestResult ReadResult      `json:"filesRequestResponse,omitzero"`
		WriteFilesResult   WriteResult     `json:"writeFilesResult,omitzero"`
	}
	Output struct {
		Text           string         `json:"text,omitempty,omitzero"`
		ExecuteRequest ExecuteRequest `json:"executeRequest,omitempty,omitzero"`
		ReadRequest    ReadRequest    `json:"readRequest,omitempty,omitzero"`
		WriteRequest   WriteRequest   `json:"writeRequest,omitempty,omitzero"`
	}

	Platform int
)

func (o Output) IsFunction() bool {
	return o.IsExecuteRequest() || o.IsReadRequest() || o.IsWriteRequest()
}

func (o Output) IsExecuteRequest() bool {
	return o.ExecuteRequest.Text != ""
}

func (o Output) IsReadRequest() bool {
	return len(o.ReadRequest.FilePaths) > 0
}

func (o Output) IsWriteRequest() bool {
	return len(o.WriteRequest.Files) > 0
}

func (i Input) IsCommandResult() bool {
	return i.ExecuteResult.Executed
}

func (i Input) IsFilesRequestResult() bool {
	return i.FilesRequestResult.FilesAttached
}
