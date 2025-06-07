package schema

import "encoding/json"

const (
	FinishReasonStop      = "STOP"
	FinishReasonMaxTokens = "MAX_TOKENS"
)

type (
	UploadedFile struct {
		File struct {
			Name        string `json:"name"`
			DisplayName string `json:"displayName"`
			MimeType    string `json:"mimeType"`
			SizeBytes   string `json:"sizeBytes"`
			CreateTime  string `json:"createTime"`
			UpdateTime  string `json:"updateTime"`
			URI         string `json:"uri"`
		} `json:"file"`
	}
	Request struct {
		SystemInstruction SystemInstruction `json:"system_instruction"`
		Contents          []Content         `json:"contents"`
		Tools             []json.RawMessage `json:"tools"`
		GenerationConfig  GenerationConfig  `json:"generationConfig"`
	}
	SystemInstruction struct {
		Parts []Part `json:"parts"`
	}
	GenerationConfig struct {
		Temperature      float64         `json:"temperature"`
		TopP             float64         `json:"topP"`
		MaxOutputTokens  int             `json:"maxOutputTokens"`
		ResponseMimeType string          `json:"responseMimeType,omitempty"`
		ResponseSchema   json.RawMessage `json:"responseSchema,omitempty"`
	}
)

type (
	Response struct {
		Candidates    []Candidate   `json:"candidates"`
		UsageMetadata UsageMetadata `json:"usageMetadata"`
	}
	UsageMetadata struct {
		TotalTokenCount int `json:"totalTokenCount"`
	}
)

type (
	Part struct {
		Text             string          `json:"text,omitzero"`
		File             *FileData       `json:"fileData,omitempty"`
		FunctionCall     FunctionCall    `json:"functionCall,omitempty,omitzero"`
		FunctionResponse json.RawMessage `json:"function_response,omitempty,omitzero"`
	}
	FileData struct {
		MIMEType string `json:"mimeType"`
		URI      string `json:"fileUri"`
	}
	FunctionCall struct {
		Name string          `json:"name"`
		Args json.RawMessage `json:"args"`
	}
	Content struct {
		Role  string `json:"role"`
		Parts []Part `json:"parts"`
	}
	Candidate struct {
		Content      Content `json:"content"`
		FinishReason string  `json:"finishReason"`
	}
)
