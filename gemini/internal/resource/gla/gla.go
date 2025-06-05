// Package gla encapsulates Google's Generative Language API resource management functionality
package gla

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/comradequinn/gen/gemini/internal/resource"
)

func Upload(uploadRequest resource.UploadRequest, debugPrintf func(msg string, args ...any)) (resource.Reference, error) {

	fileInfo, contentType, err := resource.FileInfo(uploadRequest.File)

	if err != nil {
		return resource.Reference{}, err
	}

	url := strings.ReplaceAll(uploadRequest.URL, "{api-key}", uploadRequest.Credential)

	rq, err := http.NewRequest(http.MethodPost, url, strings.NewReader(fmt.Sprintf(`{"file":{"display_name":"%v"}}`, fileInfo.Name())))

	if err != nil {
		return resource.Reference{}, fmt.Errorf("unable to create start-upload request. %w", err)
	}

	rq.Header.Set("X-Goog-Upload-Protocol", "resumable")
	rq.Header.Set("X-Goog-Upload-Command", "start")
	rq.Header.Set("X-Goog-Upload-Header-Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
	rq.Header.Set("X-Goog-Upload-Header-Content-Type", contentType)
	rq.Header.Set("Content-Type", "application/json")

	debugPrintf("sending start upload request", "type", "start_upload_request", "url", url, "headers", rq.Header)

	rs, err := http.DefaultClient.Do(rq)
	if err != nil {
		return resource.Reference{}, fmt.Errorf("error starting file upload. %w", err)
	}
	defer rs.Body.Close()

	body, _ := io.ReadAll(rs.Body)

	debugPrintf("received start upload response", "type", "start_upload_response", "status", rs.Status, "response", string(body))

	if rs.StatusCode != http.StatusOK {
		return resource.Reference{}, fmt.Errorf("start-upload request failed with status code %v. %v", rs.StatusCode, string(body))
	}

	uploadURL := rs.Header.Get("X-Goog-Upload-Url")
	if uploadURL == "" {
		return resource.Reference{}, fmt.Errorf("upload url not found in start-upload response header of 'x-goog-upload-url'")
	}

	file, err := resource.FileIO.Open(uploadRequest.File)
	if err != nil {
		return resource.Reference{}, fmt.Errorf("unable to open file '%v' for upload. %w", uploadRequest.File, err)
	}
	defer file.Close()

	rq, err = http.NewRequest(http.MethodPost, uploadURL, file) // Use the file as the request body
	if err != nil {
		return resource.Reference{}, fmt.Errorf("unable to create upload-request. %w", err)
	}

	rq.Header.Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
	rq.Header.Set("X-Goog-Upload-Offset", "0")
	rq.Header.Set("X-Goog-Upload-Command", "upload, finalize")

	debugPrintf("sending upload request", "type", "upload_request", "url", url, "headers", rq.Header, "bytes", strconv.FormatInt(fileInfo.Size(), 10))

	rs, err = http.DefaultClient.Do(rq)
	if err != nil {
		return resource.Reference{}, fmt.Errorf("error during upload-request. %w", err)
	}
	defer rs.Body.Close()

	body, err = io.ReadAll(rs.Body)

	debugPrintf("received upload response", "type", "upload_response", "status", rs.Status, "response", string(body))

	if rs.StatusCode != http.StatusOK || err != nil {
		return resource.Reference{}, fmt.Errorf("upload-request failed with status code %v. error: %w. body: %v", rs.StatusCode, err, string(body))
	}

	uploadResponse := struct {
		File struct {
			MimeType string `json:"mimeType"`
			URI      string `json:"uri"`
		} `json:"file"`
	}{}

	if err = json.Unmarshal(body, &uploadResponse); err != nil {
		return resource.Reference{}, fmt.Errorf("unable to marshal upload-request response. %w", err)
	}

	return resource.Reference{
		URI:      uploadResponse.File.URI,
		MIMEType: uploadResponse.File.MimeType,
		Label:    fileInfo.Name(),
	}, nil
}
