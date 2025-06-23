package gemini

import (
	"fmt"
	"strings"
	"testing"
)

func TestConfig_withdefaults(t *testing.T) {
	basePrompt := Prompt{}
	validBaseConfig := Config{
		MaxTokens:   100,
		Temperature: 0.7,
	}

	tests := []struct {
		name       string
		cfg        Config
		prompt     Prompt
		wantErr    bool
		wantErrMsg string
		validate   func(t *testing.T, cfg Config, originalCfg Config)
	}{
		{
			name:   "valid config - generative language defaults",
			cfg:    validBaseConfig,
			prompt: basePrompt,
			validate: func(t *testing.T, cfg Config, originalCfg Config) {
				if !strings.HasPrefix(cfg.GeminiURL, "https://generativelanguage.googleapis.com/v1beta/models/") {
					t.Errorf("expected geminiurl to have generativelanguage default, got %s", cfg.GeminiURL)
				}
				if !strings.HasPrefix(cfg.FileStorageURL, "https://generativelanguage.googleapis.com/upload/v1beta/files") {
					t.Errorf("expected filestorageurl to have generativelanguage default, got %s", cfg.FileStorageURL)
				}
				if !strings.Contains(cfg.SystemPrompt, fmt.Sprintf(". Your responses must not exceed %v words in length. ", float64(originalCfg.MaxTokens)*0.75)) {
					t.Errorf("systemprompt does not contain maxtokens constraint")
				}
			},
		},
		{
			name: "valid config - vertex defaults",
			cfg: Config{
				MaxTokens:   200,
				Temperature: 0.8,
				GCPProject:  "test-project",
				GCSBucket:   "test-bucket",
				Model:       "gemini-pro",
			},
			prompt: basePrompt,
			validate: func(t *testing.T, cfg Config, originalCfg Config) {
				expectedGeminiURL := "https://aiplatform.googleapis.com/v1/projects/test-project/locations/global/publishers/google/models/gemini-pro:generateContent"
				if cfg.GeminiURL != expectedGeminiURL {
					t.Errorf("expected geminiurl to be %s, got %s", expectedGeminiURL, cfg.GeminiURL)
				}
				expectedFileStorageURL := "https://storage.googleapis.com/upload/storage/v1/b/test-bucket/o?uploadType=media&name={file-name}"
				if cfg.FileStorageURL != expectedFileStorageURL { // {file-name} is not replaced in withdefaults
					t.Errorf("expected filestorageurl to be %s, got %s", expectedFileStorageURL, cfg.FileStorageURL)
				}
				if !strings.Contains(cfg.SystemPrompt, fmt.Sprintf(". Your responses must not exceed %v words in length. ", float64(originalCfg.MaxTokens)*0.75)) {
					t.Errorf("systemprompt does not contain maxtokens constraint")
				}
			},
		},
		{
			name: "invalid config - maxtokens zero",
			cfg: Config{
				Temperature: 1.0,
			},
			prompt:     basePrompt,
			wantErr:    true,
			wantErrMsg: "invalid configuration. maxtokens must be specified",
		},
		{
			name: "invalid config - schema with execution",
			cfg: Config{
				MaxTokens:        100,
				Temperature:      0.7,
				ExecutionEnabled: true,
			},
			prompt: Prompt{
				Schema: "some-schema",
			},
			wantErr:    true,
			wantErrMsg: "invalid prompt or configuration. a response schema cannot be specified when execution is enabled",
		},
		{
			name: "invalid config - gcpproject without gcsbucket",
			cfg: Config{
				MaxTokens:   100,
				Temperature: 0.7,
				GCPProject:  "test-project",
			},
			prompt:     basePrompt,
			wantErr:    true,
			wantErrMsg: "to use the gemini api via vertex-ai a gcp-project, gcs-bucket and vertex-access-token must be provided",
		},
		{
			name: "invalid config - gcsbucket without gcpproject",
			cfg: Config{
				MaxTokens:   100,
				Temperature: 0.7,
				GCSBucket:   "test-bucket",
			},
			prompt:     basePrompt,
			wantErr:    true,
			wantErrMsg: "to use the gemini api via vertex-ai a gcp-project, gcs-bucket and vertex-access-token must be provided",
		},
		{
			name: "grounding disabled - with schema",
			cfg: Config{
				MaxTokens:   100,
				Temperature: 0.7,
				Grounding:   true,
			},
			prompt: Prompt{
				Schema: "some-schema",
			},
			validate: func(t *testing.T, cfg Config, originalCfg Config) {
				if cfg.Grounding {
					t.Error("expected grounding to be false when schema is provided, got true")
				}
			},
		},
		{
			name: "grounding disabled - with execution",
			cfg: Config{
				MaxTokens:        100,
				Temperature:      0.7,
				Grounding:        true,
				ExecutionEnabled: true,
			},
			prompt: basePrompt,
			validate: func(t *testing.T, cfg Config, originalCfg Config) {
				if cfg.Grounding {
					t.Error("expected grounding to be false when execution is true, got true")
				}
			},
		},
		{
			name: "custom urls provided",
			cfg: Config{
				MaxTokens:      100,
				Temperature:    0.7,
				GeminiURL:      "custom-gemini/{model}",
				FileStorageURL: "custom-storage/{api-key}", // api-key is not replaced by formaturl
				Model:          "test-model",
			},
			prompt: basePrompt,
			validate: func(t *testing.T, cfg Config, originalCfg Config) {
				if cfg.GeminiURL != "custom-gemini/test-model" {
					t.Errorf("expected custom geminiurl to be 'custom-gemini/test-model', got %s", cfg.GeminiURL)
				}
				if cfg.FileStorageURL != "custom-storage/{api-key}" {
					t.Errorf("expected custom filestorageurl to be 'custom-storage/{api-key}', got %s", cfg.FileStorageURL)
				}
			},
		},
		{
			name: "system prompt with use case",
			cfg: Config{
				MaxTokens:    100,
				Temperature:  0.7,
				SystemPrompt: "Base prompt.",
				UseCase:      "Test use case.",
			},
			prompt: basePrompt,
			validate: func(t *testing.T, cfg Config, originalCfg Config) {
				expectedPrompt := "Base prompt.. Your responses must not exceed 75 words in length. Consider in your responses, where it may be relevant, that the following information has been provided about your specific use-case: [Test use case.]"
				if cfg.SystemPrompt != expectedPrompt {
					t.Errorf("expected systemprompt to be '%s', got '%s'", expectedPrompt, cfg.SystemPrompt)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalCfg := tt.cfg // keep a copy for validation
			gotCfg, err := tt.cfg.withDefaults(tt.prompt)
			if (err != nil) != tt.wantErr {
				t.Fatalf("withdefaults() error = %v, wanterr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err.Error() != tt.wantErrMsg {
					t.Fatalf("withdefaults() error msg = %q, wanterrmsg %q", err.Error(), tt.wantErrMsg)
				}
				return
			}
			if tt.validate != nil {
				tt.validate(t, gotCfg, originalCfg)
			}
		})
	}
}
