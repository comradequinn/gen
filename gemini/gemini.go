// Package gemini defines an interface to the Gemini API
package gemini

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/comradequinn/gen/gemini/internal/resource"
	"github.com/comradequinn/gen/gemini/internal/resource/gcs"
	"github.com/comradequinn/gen/gemini/internal/resource/gla"
	"github.com/comradequinn/gen/gemini/internal/schema"
	"github.com/comradequinn/gen/log"
)

const (
	PlatformVertex             Platform = 1
	PlatformGenerativeLanguage Platform = 2
)

var (
	Models = struct {
		Pro   string
		Flash string
	}{
		Pro:   "gemini-2.5-pro-preview-06-05",
		Flash: "gemini-2.5-flash-preview-05-20",
	}
)

const (
	InputTypeUser    = "user"
	InputTypeCommand = "command"
)

const (
	RoleUser  = "user"
	RoleModel = "model"
	RoleTool  = "tool"
)

// Generate queries the Gemini API with the specified prompt and returns the result
func Generate(cfg Config, prompt Prompt) (Transaction, error) {
	var err error

	if cfg, err = cfg.withDefaults(prompt); err != nil {
		return Transaction{}, fmt.Errorf("invalid configuration. %w", err)
	}

	contents := historyFrom(prompt)

	var (
		part                schema.Part
		resourceRefs        []resource.Reference
		resourceUploadFunc  resource.UploadFunc
		url                 string
		authorisationHeader string
		role                string
	)

	switch {
	case prompt.InputType == InputTypeUser:
		part = schema.Part{Text: prompt.Text}
		role = RoleUser
	case prompt.CommandResult.Executed:
		part = schema.Part{FunctionResponse: prompt.CommandResult.marshalJSON()}
		role = RoleUser
	case prompt.FilesRequestResult.Attached:
		part = schema.Part{FunctionResponse: prompt.FilesRequestResult.marshalJSON()}
		role = RoleUser
	}

	content := schema.Content{
		Role:  role,
		Parts: []schema.Part{part},
	}

	switch cfg.platform() {
	case PlatformGenerativeLanguage:
		resourceUploadFunc = gla.Upload
		url = strings.ReplaceAll(cfg.GeminiURL, "{api-key}", cfg.Credential)
	case PlatformVertex:
		resourceUploadFunc = gcs.Upload
		url = cfg.GeminiURL
		authorisationHeader = "Bearer " + cfg.Credential
	default:
		panic(fmt.Sprintf("unsupported api platform %v", cfg.platform()))
	}

	if len(prompt.Files) > 0 {
		if resourceRefs, err = resource.Upload(resource.BatchUploadRequest{
			URL:        cfg.FileStorageURL,
			Credential: cfg.Credential,
			UploadFunc: resourceUploadFunc,
			Files:      prompt.Files,
		}); err != nil {
			return Transaction{}, err
		}

		for _, ref := range resourceRefs {
			content.Parts = append(content.Parts, schema.Part{File: &schema.FileData{URI: ref.URI, MIMEType: ref.MIMEType}})
		}
	}

	contents = append(contents, content)

	tools := []json.RawMessage{}

	if cfg.Grounding {
		tools = append(tools, googleSearchTool{}.marshalJSON())
	}

	if cfg.CommandExecution {
		tools = append(tools, commandExecutionTool{}.marshalJSON())
	}

	generationConfig := schema.GenerationConfig{
		Temperature:      cfg.Temperature,
		TopP:             cfg.TopP,
		MaxOutputTokens:  cfg.MaxTokens,
		ResponseMimeType: "text/plain",
	}

	if prompt.Schema != "" {
		generationConfig.ResponseMimeType = "application/json"
		generationConfig.ResponseSchema = json.RawMessage(prompt.Schema)
	}

	response, err := geminiHTTP(url, authorisationHeader, cfg, contents, tools, generationConfig)

	if err != nil {
		return Transaction{}, err
	}

	responseText, commandRequest, filesRequest := strings.Builder{}, CommandRequest{}, FilesRequest{}

	for _, part := range response.Candidates[0].Content.Parts {
		if part.FunctionCall.Name != "" {
			switch {
			case part.FunctionCall.Name == (commandExecutionTool{}).ExecCmdFunctionName():
				if err := json.NewDecoder(bytes.NewReader(part.FunctionCall.Args)).Decode(&commandRequest); err != nil {
					return Transaction{}, fmt.Errorf("unable to decode function call arguments for '%v' returned from gemini api. %w", part.FunctionCall.Name, err)
				}
			case part.FunctionCall.Name == (commandExecutionTool{}).RequestFilesFunctionName():
				if err := json.NewDecoder(bytes.NewReader(part.FunctionCall.Args)).Decode(&filesRequest); err != nil {
					return Transaction{}, fmt.Errorf("unable to decode function call arguments for '%v' returned from gemini api. %w", part.FunctionCall.Name, err)
				}
			default:
				return Transaction{}, fmt.Errorf("unexpected function call response returned from gemini api. zero or one function of types '%v' or '%v' expected. got %+v", (commandExecutionTool{}).ExecCmdFunctionName(), (commandExecutionTool{}).RequestFilesFunctionName(), part.FunctionCall)
			}
		}

		responseText.WriteString(part.Text)
	}

	log.DebugPrintf("token count value reported", "type", "report", "token_count", response.UsageMetadata.TotalTokenCount)

	filesReferences := make([]FileReference, 0, len(resourceRefs))

	for _, resourceRef := range resourceRefs {
		filesReferences = append(filesReferences, FileReference{
			URI:      resourceRef.URI,
			MIMEType: resourceRef.MIMEType,
			Label:    resourceRef.Label,
		})
	}

	transaction := Transaction{
		Tokens: response.UsageMetadata.TotalTokenCount,
		Input: Input{
			Type:           prompt.InputType,
			Text:           prompt.Text,
			CommandResult:  prompt.CommandResult,
			FileReferences: filesReferences,
		},
		Output: Output{
			Text:           responseText.String(),
			CommandRequest: commandRequest,
			FilesRequest:   filesRequest,
		},
	}

	if transaction.Output.IsFunction() {
		transaction.Output.Text = "" // discard any, typically inconsistent, output describing command execution, it can be derived from the function itself more consistently
	}

	if transaction.Output.Text == "" && !transaction.Output.IsFunction() {
		return Transaction{}, fmt.Errorf("unexpected text response returned from gemini api. expected text content. got empty string")
	}

	return transaction, nil
}
