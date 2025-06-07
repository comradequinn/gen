package gemini

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/comradequinn/gen/gemini/internal/schema"
	"github.com/comradequinn/gen/log"
)

func geminiHTTP(url, authorisationHeader string, cfg Config, contents []schema.Content, tools []json.RawMessage, generationConfig schema.GenerationConfig) (schema.Response, error) {
	request := bytes.Buffer{}
	if err := json.NewEncoder(&request).Encode(schema.Request{
		SystemInstruction: schema.SystemInstruction{
			Parts: []schema.Part{{Text: cfg.SystemPrompt}},
		},
		Contents:         contents,
		Tools:            tools,
		GenerationConfig: generationConfig,
	}); err != nil {
		return schema.Response{}, fmt.Errorf("unable to encode gemini request as json. %w", err)
	}

	rq, _ := http.NewRequest(http.MethodPost, url, &request)
	rq.Header.Set("Content-Type", "application/json")
	rq.Header.Set("Authorization", authorisationHeader)

	log.DebugPrintf("sending generate request", "type", "generate_request", "url", url, "headers", rq.Header, "body", request.String())

	rs, err := http.DefaultClient.Do(rq)

	if err != nil {
		return schema.Response{}, fmt.Errorf("unable to send request to gemini api. %w", err)
	}

	defer rs.Body.Close()

	body, err := io.ReadAll(rs.Body)

	if err != nil {
		return schema.Response{}, fmt.Errorf("unable to read response body. %w", err)
	}

	log.DebugPrintf("received generate response", "type", "generate_response", "status", rs.Status, "request", string(body))

	if rs.StatusCode != http.StatusOK {
		return schema.Response{}, fmt.Errorf("non-200 status code returned from gemini api. %s", body)
	}

	response := schema.Response{}

	if err := json.Unmarshal(body, &response); err != nil || len(response.Candidates) == 0 {
		return schema.Response{}, fmt.Errorf("unable to parse response body or no valid response candidates returned. response: [%s]. error: %w", string(body), err)
	}

	switch response.Candidates[0].FinishReason {
	case schema.FinishReasonStop:
	case schema.FinishReasonMaxTokens:
		return schema.Response{}, fmt.Errorf("the response was terminated before it completed as the maximum number of tokens was reached")
	default:
		return schema.Response{}, fmt.Errorf("the response was terminated before it completed. the stated reason was '%v'", response.Candidates[0].FinishReason)
	}

	return response, nil
}
