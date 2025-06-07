package gemini

import (
	"fmt"
	"strings"
)

type (
	Config struct {
		GeminiURL        string
		FileStorageURL   string
		Credential       string
		GCPProject       string
		GCSBucket        string
		Model            string
		SystemPrompt     string
		MaxTokens        int
		Temperature      float64
		TopP             float64
		UseCase          string
		Grounding        bool
		CommandExecution bool
		CommandApproval  bool
	}
)

func (cfg Config) platform() Platform {
	if cfg.GCPProject != "" {
		return PlatformVertex
	}

	return PlatformGenerativeLanguage
}

func (cfg Config) withDefaults(prompt Prompt) (Config, error) {
	if cfg.MaxTokens == 0 {
		return cfg, fmt.Errorf("invalid configuration. maxtokens must be specified")
	}

	if prompt.Schema != "" && cfg.CommandExecution {
		return cfg, fmt.Errorf("invalid prompt or configuration. a response schema cannot be specified when command-execution is enabled")
	}

	if (cfg.GCPProject != "" || cfg.GCSBucket != "") && (cfg.GCPProject == "" || cfg.GCSBucket == "") {
		return cfg, fmt.Errorf("to use the gemini api via vertex-ai a gcp-project, gcs-bucket and vertex-access-token must be provided")
	}

	if (prompt.Schema != "" || cfg.CommandExecution) && cfg.Grounding {
		cfg.Grounding = false
	}

	switch cfg.platform() {
	case PlatformGenerativeLanguage:
		if cfg.GeminiURL == "" {
			cfg.GeminiURL = "https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent?key={api-key}"
		}
		if cfg.FileStorageURL == "" {
			cfg.FileStorageURL = "https://generativelanguage.googleapis.com/upload/v1beta/files?key={api-key}"
		}
	case PlatformVertex:
		if cfg.GeminiURL == "" {
			cfg.GeminiURL = "https://aiplatform.googleapis.com/v1/projects/{gcp-project}/locations/global/publishers/google/models/{model}:generateContent"
		}
		if cfg.FileStorageURL == "" {
			cfg.FileStorageURL = "https://storage.googleapis.com/upload/storage/v1/b/{gcs-bucket}/o?uploadType=media&name={file-name}"
		}
	}

	formatURL := func(u string) string {
		return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(u,
			"{model}", cfg.Model),
			"{gcp-project}", cfg.GCPProject),
			"{gcs-bucket}", cfg.GCSBucket)
	}

	cfg.GeminiURL, cfg.FileStorageURL = formatURL(cfg.GeminiURL), formatURL(cfg.FileStorageURL)

	cfg.SystemPrompt += fmt.Sprintf(". Your responses must not exceed %v words in length. ", float64(cfg.MaxTokens)*0.75)

	if cfg.UseCase != "" {
		cfg.SystemPrompt += "Consider in your responses, where it may be relevant, that the following information has been provided about your specific use-case: [" + cfg.UseCase + "]"
	}

	return cfg, nil
}
