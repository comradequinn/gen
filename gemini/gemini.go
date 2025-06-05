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
	}
	Prompt struct {
		History []Message
		Text    string
		Files   []string
		Schema  string
	}
	FileReference struct {
		URI      string `json:"uri"`
		MIMEType string `json:"mimeType"`
		Label    string `json:"label"`
	}
	Response struct {
		Tokens int
		Text   string
		Files  []FileReference
	}
	Role    string
	Message struct {
		Role  Role            `json:"role"`
		Text  string          `json:"text"`
		Files []FileReference `json:"files,omitempty"`
	}
	Platform int
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
	RoleUser  = "user"
	RoleModel = "model"
)

// Generate queries the Gemini API with the specified prompt and returns the result
func Generate(cfg Config, prompt Prompt, debugPrintf func(msg string, args ...any)) (Response, error) {
	if cfg.MaxTokens == 0 || cfg.Temperature == 0 {
		return Response{}, fmt.Errorf("invalid prompt. maxtokens and temperature must be specified")
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

	for _, message := range prompt.History {
		content := schema.Content{
			Role:  string(message.Role),
			Parts: []schema.Part{{Text: message.Text}}}
		if len(message.Files) > 0 {
			for _, fileReference := range message.Files {
				content.Parts = append(content.Parts, schema.Part{
					File: &schema.FileData{URI: fileReference.URI, MIMEType: fileReference.MIMEType},
				})
			}
		}

		contents = append(contents, content)
	}

	content := schema.Content{
		Role: RoleUser,
		Parts: []schema.Part{
			{Text: prompt.Text},
		},
	}

	var (
		resourceRefs        []resource.Reference
		resourceUploadFunc  resource.UploadFunc
		url                 string
		authorisationHeader string
		err                 error
	)

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
			return Response{}, err
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
		return Response{}, fmt.Errorf("unable to encode gemini request as json. %w", err)
	}

	rq, _ := http.NewRequest(http.MethodPost, url, &request)

	rq.Header.Set("Content-Type", "application/json")
	rq.Header.Set("Authorization", authorisationHeader)

	debugPrintf("sending generate request", "type", "generate_request", "url", url, "headers", rq.Header, "body", request.String())

	rs, err := http.DefaultClient.Do(rq)

	if err != nil {
		return Response{}, fmt.Errorf("unable to send request to gemini api. %w", err)
	}

	defer rs.Body.Close()

	body, err := io.ReadAll(rs.Body)

	if err != nil {
		return Response{}, fmt.Errorf("unable to read response body. %w", err)
	}

	debugPrintf("received generate response", "type", "generate_response", "status", rs.Status, "request", string(body))

	if rs.StatusCode != 200 {
		return Response{}, fmt.Errorf("non-200 status code returned from gemini api. %s", body)
	}

	response := schema.Response{}

	if err := json.Unmarshal(body, &response); err != nil {
		return Response{}, fmt.Errorf("unable to parse response body. %w", err)
	}

	if len(response.Candidates) == 0 || response.Candidates[0].FinishReason != schema.FinishReasonStop {
		return Response{}, fmt.Errorf("no valid response candidates returned. response: %s", body)
	}

	sb := strings.Builder{}

	for _, part := range response.Candidates[0].Content.Parts {
		sb.WriteString(part.Text)
	}

	debugPrintf("token count value reported", "type", "report", "token_count", response.UsageMetadata.TotalTokenCount)

	files := make([]FileReference, 0, len(resourceRefs))

	for _, resourceRef := range resourceRefs {
		files = append(files, FileReference{
			URI:      resourceRef.URI,
			MIMEType: resourceRef.MIMEType,
			Label:    resourceRef.Label,
		})
	}

	return Response{
		Tokens: response.UsageMetadata.TotalTokenCount,
		Text:   sb.String(),
		Files:  files,
	}, nil
}
