// Package gcs encapsulates Google Cloud Storage resource management functionality
package gcs

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/comradequinn/gen/gemini/internal/resource"
)

func Upload(uploadRequest resource.UploadRequest, debugPrintf func(msg string, args ...any)) (resource.Reference, error) {

	fileInfo, contentType, err := resource.FileInfo(uploadRequest.File)

	if err != nil {
		return resource.Reference{}, err
	}

	file, err := resource.FileIO.Open(uploadRequest.File)

	if err != nil {
		return resource.Reference{}, fmt.Errorf("unable to open file '%v' for upload. %w", uploadRequest.File, err)
	}

	defer file.Close()

	url := strings.ReplaceAll(uploadRequest.URL, "{file-name}", url.QueryEscape(fmt.Sprintf("gen-attachment-%v-%v-%v", fileInfo.Name(), strconv.FormatInt(time.Now().UnixNano(), 10), strconv.Itoa(rand.Int()))))

	rq, err := http.NewRequest(http.MethodPost, url, file)
	if err != nil {
		return resource.Reference{}, fmt.Errorf("unable to create upload-request. %w", err)
	}

	rq.Header.Set("Content-Type", contentType)
	rq.Header.Set("Authorization", "Bearer "+uploadRequest.Credential)

	debugPrintf("sending upload request", "type", "upload_request", "url", url, "headers", rq.Header, "bytes", strconv.FormatInt(fileInfo.Size(), 10))

	rs, err := http.DefaultClient.Do(rq)
	if err != nil {
		return resource.Reference{}, fmt.Errorf("error during upload-request. %w", err)
	}
	defer rs.Body.Close()

	body, err := io.ReadAll(rs.Body)

	if err != nil {
		return resource.Reference{}, fmt.Errorf("unable to read response body. %w", err)
	}

	debugPrintf("received upload response", "type", "upload_response", "status", rs.Status, "response", string(body))

	if rs.StatusCode != http.StatusOK || err != nil {
		return resource.Reference{}, fmt.Errorf("upload-request failed with status code %v. error: %w. body: %v", rs.StatusCode, err, string(body))
	}

	uploadResponse := struct {
		Name        string `json:"name"`
		Bucket      string `json:"bucket"`
		ContentType string `json:"contentType"`
	}{}

	if err = json.Unmarshal(body, &uploadResponse); err != nil {
		return resource.Reference{}, fmt.Errorf("unable to marshal upload-request response. %w", err)
	}

	return resource.Reference{
		URI:      fmt.Sprintf("gs://%v/%v", uploadResponse.Bucket, uploadResponse.Name),
		MIMEType: uploadResponse.ContentType,
		Label:    fileInfo.Name(),
	}, nil
}
