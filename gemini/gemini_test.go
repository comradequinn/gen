package gemini_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/comradequinn/gen/gemini"
	"github.com/comradequinn/gen/gemini/internal/resource"
	"github.com/comradequinn/gen/gemini/internal/schema"
	"github.com/comradequinn/gen/log"
)

type MockFileInfo struct{ name string }

func (m MockFileInfo) Name() string       { return m.name }
func (m MockFileInfo) Size() int64        { return 10 }
func (m MockFileInfo) Mode() os.FileMode  { return 0 }
func (m MockFileInfo) ModTime() time.Time { return time.Time{} }
func (m MockFileInfo) IsDir() bool        { return false }
func (m MockFileInfo) Sys() any           { return nil }

func TestGenerate(t *testing.T) {
	log.Init(false, func(string, ...any) {})

	resource.FileIO.Stat = func(name string) (os.FileInfo, error) {
		return MockFileInfo{
			name: name,
		}, nil
	}
	resource.FileIO.Open = func(name string) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("test-data")), nil
	}

	expectedFileURI := "test-file-ref-uri"
	actualRq := schema.Request{}
	expectedResponse := schema.Response{
		Candidates: []schema.Candidate{
			{
				Content: schema.Content{
					Role: "model",
					Parts: []schema.Part{
						{Text: "test-response-a"},
						{Text: "test-response-b"},
					},
				},
				FinishReason: schema.FinishReasonStop,
			},
			{
				Content: schema.Content{
					Role: "model",
					Parts: []schema.Part{
						{Text: "test-response-ignore-a"},
						{Text: "test-response-ignore-b"},
					},
				},
				FinishReason: schema.FinishReasonStop,
			},
		},
		UsageMetadata: schema.UsageMetadata{
			TotalTokenCount: 1000,
		},
	}

	var svr *httptest.Server
	svr = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/test-generate-url/":
			if r.URL.Query().Get("api-key") != "test=api-key" {
				t.Fatalf("expected api key to be %v. got %v", "test=api-key", r.URL.Query()["api-key"][0])
			}

			if err := json.NewDecoder(r.Body).Decode(&actualRq); err != nil {
				t.Fatalf("unable to decode gemini stub request body. %v", err)
			}

			if err := json.NewEncoder(w).Encode(&expectedResponse); err != nil {
				t.Fatalf("unable to encode gemini stub response body. %v", err)
			}
		case r.URL.Path == "/test-start-upload-url/":
			w.Header().Set("X-Goog-Upload-Url", svr.URL+"/test-upload-url/")
		case r.URL.Path == "/test-upload-url/":
			rs := map[string]interface{}{
				"file": map[string]string{
					"displayName": "test-file-1",
					"mimeType":    "text/plain",
					"uri":         expectedFileURI,
				},
			}
			json.NewEncoder(w).Encode(rs)
		default:
			t.Fatalf("unexpected request to %v", r.URL.Path)
		}
	}))
	defer svr.Close()

	cfg := gemini.Config{
		Credential:     "test=api-key",
		GeminiURL:      svr.URL + "/test-generate-url/?model=gemini&api-key={api-key}",
		FileStorageURL: svr.URL + "/test-start-upload-url/?api-key=%v",
		SystemPrompt:   "test-system-prompt",
		MaxTokens:      1000,
		Temperature:    1.0,
		TopP:           1.0,
		Grounding:      true,
		UseCase:        "you are helping a test-user in test-location with test-description",
	}

	prompt := gemini.Prompt{
		Text:      "test prompt",
		InputType: gemini.InputTypeUser,
		History: []gemini.Transaction{
			{
				Input: gemini.Input{
					Type: gemini.InputTypeUser,
					Text: "test-history-1",
				},
				Output: gemini.Output{
					Text: "test-history-2",
				},
			},
		},
		Files: []string{
			"test-file-1",
		},
	}

	assert := func(t *testing.T, condition bool, format string, v ...any) {
		if !condition {
			t.Fatalf(format, v...)
		}
	}

	assertResponse := func(t *testing.T, rs gemini.Transaction, err error) {
		assert(t, err == nil, "expected no error generating response. got %v", err)
		assert(t, actualRq.GenerationConfig.MaxOutputTokens == cfg.MaxTokens, "expected max output tokens to be %v. got %v", cfg.MaxTokens, actualRq.GenerationConfig.MaxOutputTokens)
		assert(t, actualRq.GenerationConfig.Temperature == cfg.Temperature, "expected temperature to be %v. got %v", cfg.Temperature, actualRq.GenerationConfig.Temperature)
		assert(t, actualRq.GenerationConfig.TopP == cfg.TopP, "expected top-p to be %v. got %v", cfg.TopP, actualRq.GenerationConfig.TopP)

		if cfg.Grounding {
			assert(t, len(actualRq.Tools) == 1 && strings.Contains(string(actualRq.Tools[0]), "googleSearch"), "expected 1 tool of type google-search to be specified when grounding enabled. got %v", len(actualRq.Tools))
		} else {
			assert(t, len(actualRq.Tools) == 0, "expected 0 tools to be specified when grounding disabled. got %v", len(actualRq.Tools))
		}

		if string(prompt.Schema) != "" {
			assert(t, actualRq.GenerationConfig.ResponseMimeType == "application/json", "expected response mime type to be application/json when a response schema is specified. got %v")
			data, _ := actualRq.GenerationConfig.ResponseSchema.MarshalJSON()
			assert(t, string(data) == string(prompt.Schema), "expected response schema to be %v. got %v", prompt.Schema, string(data))
		} else {
			assert(t, actualRq.GenerationConfig.ResponseMimeType == "text/plain", "expected response mime type to be text/plain when no response schema specified. got %v", actualRq.GenerationConfig.ResponseMimeType)
		}

		systemPrompt := actualRq.SystemInstruction.Parts[0].Text

		assert(t, strings.Contains(systemPrompt, cfg.SystemPrompt), "expected system prompt %q to contain %q", systemPrompt, cfg.SystemPrompt)
		assert(t, strings.Contains(systemPrompt, cfg.UseCase), "expected system prompt to contain %v", cfg.UseCase)
		assert(t, len(actualRq.Contents) == 3, "expected 3 content entries. got %v", len(actualRq.Contents))

		for i := range len(prompt.History) {
			assert(t, actualRq.Contents[i].Role == string(prompt.History[i].Input.Type), "expected input type to be %v. got %v", gemini.InputTypeUser, actualRq.Contents[0].Role)
			assert(t, actualRq.Contents[i].Parts[0].Text == string(prompt.History[i].Input.Text), "expected input text to be %v. got %v", prompt.Text, actualRq.Contents[0].Parts[0].Text)
		}

		assert(t, actualRq.Contents[2].Role == string(gemini.RoleUser), "expected role to be %v. got %v", gemini.RoleUser, actualRq.Contents[2].Role)
		assert(t, actualRq.Contents[2].Parts[0].Text == prompt.Text, "expected text to be %v. got %v", prompt.Text, actualRq.Contents[2].Parts[0].Text)
		assert(t, actualRq.Contents[2].Parts[1].File.URI == expectedFileURI, "expected file uri to be %v. got %v", expectedFileURI, actualRq.Contents[2].Parts[1].File.URI)
		assert(t, rs.Output.Text == expectedResponse.Candidates[0].Content.Parts[0].Text+expectedResponse.Candidates[0].Content.Parts[1].Text, "expected response text to be %v. got %v", expectedResponse.Candidates[0].Content.Parts[0].Text+expectedResponse.Candidates[0].Content.Parts[1].Text, rs.Output.Text)
		assert(t, rs.Tokens == expectedResponse.UsageMetadata.TotalTokenCount, "expected response token count to be %v. got %v", expectedResponse.UsageMetadata.TotalTokenCount, rs.Tokens)
	}

	rs, err := gemini.Generate(cfg, prompt)

	assertResponse(t, rs, err)

	cfg.Grounding = false
	prompt.Schema = `{"type":"object","properties":{"response":{"type":"string"}}}`

	rs, err = gemini.Generate(cfg, prompt)

	assertResponse(t, rs, err)
}
