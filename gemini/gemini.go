package gemini

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/comradequinn/gen/gemini/internal/resource"
	"github.com/comradequinn/gen/gemini/internal/resource/gcs"
	"github.com/comradequinn/gen/gemini/internal/resource/gla"
	"github.com/comradequinn/gen/gemini/internal/schema"
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
		Pro:   "gemini-2.5-pro-preview-05-06",
		Flash: "gemini-2.5-flash-preview-04-17",
	}
)

const (
	InputTypeUser  = "user"
	InputTypeShell = "shell"
)

const (
	RoleUser  = "user"
	RoleModel = "model"
)

// Generate queries the Gemini API with the specified prompt and returns the result
func Generate(cfg Config, prompt Prompt, debugPrintf func(msg string, args ...any)) (Transaction, error) {
	if cfg.MaxTokens == 0 || cfg.Temperature == 0 {
		return Transaction{}, fmt.Errorf("invalid prompt. maxtokens and temperature must be specified")
	}

	if prompt.Schema != "" && cfg.Grounding {
		debugPrintf("grounding was specified but silently disabled due to the specification of a schema. the gemini api will not currently perform grounding for prompts requiring a structured response")
		cfg.Grounding = false
	}

	systemPrompt := strings.Builder{}
	systemPrompt.WriteString(cfg.SystemPrompt + ". ")
	systemPrompt.WriteString(fmt.Sprintf("Your responses must not exceed %v words in length. ", float64(cfg.MaxTokens)*0.75)) // rough mapping of tokens to words

	if cfg.UseCase != "" {
		systemPrompt.WriteString("Consider in your responses, where it may be relevant, that the following information has been provided about your specific use-case: [" + cfg.UseCase + "]")
	}

	contents := make([]schema.Content, 0, len(prompt.History)+1)

	for _, transaction := range prompt.History {
		content := schema.Content{
			Role:  RoleUser,
			Parts: []schema.Part{{Text: transaction.Input.Text}}}
		if len(transaction.Input.FileReferences) > 0 {
			for _, fileReference := range transaction.Input.FileReferences {
				content.Parts = append(content.Parts, schema.Part{
					File: &schema.FileData{URI: fileReference.URI, MIMEType: fileReference.MIMEType},
				})
			}
		}
		contents = append(contents, content)
		contents = append(contents, schema.Content{
			Role:  RoleModel,
			Parts: []schema.Part{{Text: transaction.Output.Text}}})
	}

	var (
		part                schema.Part
		resourceRefs        []resource.Reference
		resourceUploadFunc  resource.UploadFunc
		url                 string
		authorisationHeader string
		err                 error
	)

	if prompt.InputType == InputTypeUser {
		part = schema.Part{Text: prompt.Text}
	} else {
		j, err := prompt.ShellCommandResult.marshalJSON()

		if err != nil {
			return Transaction{}, fmt.Errorf("unable to marshal command result into json. %w", err)
		}

		part = schema.Part{FunctionResponse: json.RawMessage(j)}
	}

	content := schema.Content{
		Role:  RoleUser,
		Parts: []schema.Part{part},
	}

	switch cfg.Platform {
	case PlatformGenerativeLanguage:
		resourceUploadFunc = gla.Upload
		url = strings.ReplaceAll(cfg.GeminiURL, "{api-key}", cfg.Credential)
	case PlatformVertex:
		resourceUploadFunc = gcs.Upload
		url = cfg.GeminiURL
		authorisationHeader = "Bearer " + cfg.Credential
	default:
		panic(fmt.Sprintf("unsupported api platform %v", cfg.Platform))
	}

	if len(prompt.Files) > 0 {
		if resourceRefs, err = resource.Upload(resource.BatchUploadRequest{
			URL:        cfg.FileStorageURL,
			Credential: cfg.Credential,
			UploadFunc: resourceUploadFunc,
			Files:      prompt.Files,
		}, debugPrintf); err != nil {
			return Transaction{}, err
		}

		for _, ref := range resourceRefs {
			content.Parts = append(content.Parts, schema.Part{File: &schema.FileData{URI: ref.URI, MIMEType: ref.MIMEType}})
		}
	}

	contents = append(contents, content)

	tools := []schema.Tool{}

	if cfg.Grounding {
		tools = []schema.Tool{
			{GoogleSearch: &schema.GoogleSearch{}},
		}
	}

	if cfg.Agentic {
		shellFunction := "executes a command in the terminal of the user. the name parameter states the command to be executed. the arguments object defines each argument. " +
			"this command is only to be used to perform local operations. such as querying or interacting with the file system or a local git repo. " +
			"it is never be used as a proxy to access remote functionality. for example, using curl to invoke a http api is not an appropriate use of the command."
		tools = []schema.Tool{
			{FunctionDeclarations: []json.RawMessage{json.RawMessage(fmt.Sprintf(`{ "name": "local-shell-command","description": "%v", "parameters": `+
				`{ "type": "object", "properties": { "name":  { "type": "string" }, "arguments": { "type": "array", "items": { "type": "string" } } } } }`, shellFunction))}},
		}
	}

	generationConfig := schema.GenerationConfig{
		Temperature:     cfg.Temperature,
		TopP:            cfg.TopP,
		MaxOutputTokens: cfg.MaxTokens,
	}

	generationConfig.ResponseMimeType = "text/plain"

	if prompt.Schema != "" {
		generationConfig.ResponseMimeType = "application/json"
		generationConfig.ResponseSchema = json.RawMessage(prompt.Schema)
	}

	request := bytes.Buffer{}
	if err := json.NewEncoder(&request).Encode(schema.Request{
		SystemInstruction: schema.SystemInstruction{
			Parts: []schema.Part{{Text: systemPrompt.String()}},
		},
		Contents:         contents,
		Tools:            tools,
		GenerationConfig: generationConfig,
	}); err != nil {
		return Transaction{}, fmt.Errorf("unable to encode gemini request as json. %w", err)
	}

	rq, _ := http.NewRequest(http.MethodPost, url, &request)

	rq.Header.Set("Content-Type", "application/json")
	rq.Header.Set("Authorization", authorisationHeader)

	debugPrintf("sending generate request", "type", "generate_request", "url", url, "headers", rq.Header, "body", request.String())

	rs, err := http.DefaultClient.Do(rq)

	if err != nil {
		return Transaction{}, fmt.Errorf("unable to send request to gemini api. %w", err)
	}

	defer rs.Body.Close()

	body, err := io.ReadAll(rs.Body)

	if err != nil {
		return Transaction{}, fmt.Errorf("unable to read response body. %w", err)
	}

	debugPrintf("received generate response", "type", "generate_response", "status", rs.Status, "request", string(body))

	if rs.StatusCode != 200 {
		return Transaction{}, fmt.Errorf("non-200 status code returned from gemini api. %s", body)
	}

	response := schema.Response{}

	if err := json.Unmarshal(body, &response); err != nil {
		return Transaction{}, fmt.Errorf("unable to parse response body. %w", err)
	}

	if len(response.Candidates) == 0 || (response.Candidates[0].FinishReason != schema.FinishReasonStop && response.Candidates[0].FinishReason != schema.FinishReasonToolCall) {
		return Transaction{}, fmt.Errorf("no valid response candidates returned. response: %s", body)
	}

	sb := strings.Builder{}

	for _, part := range response.Candidates[0].Content.Parts {
		sb.WriteString(part.Text)
	}

	debugPrintf("token count value reported", "type", "report", "token_count", response.UsageMetadata.TotalTokenCount)

	filesReferences := make([]FileReference, 0, len(resourceRefs))

	for _, resourceRef := range resourceRefs {
		filesReferences = append(filesReferences, FileReference{
			URI:      resourceRef.URI,
			MIMEType: resourceRef.MIMEType,
			Label:    resourceRef.Label,
		})
	}

	return Transaction{
		Tokens: response.UsageMetadata.TotalTokenCount,
		Input: Input{
			Type:               prompt.InputType,
			Text:               prompt.Text,
			ShellCommandResult: prompt.ShellCommandResult,
			FileReferences:     filesReferences,
		},
		Output: Output{
			Text: sb.String(),
		},
	}, nil
}
