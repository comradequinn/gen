package resource

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/comradequinn/gen/log"
)

type (
	BatchUploadRequest struct {
		URL        string
		Credential string
		UploadFunc UploadFunc
		Files      []string
	}
	UploadRequest struct {
		URL        string
		Credential string
		File       string
	}
	Reference struct {
		URI      string
		MIMEType string
		Label    string
	}
	UploadFunc func(uploadRequest UploadRequest) (Reference, error)
)

var (
	FileIO = struct {
		Stat func(name string) (os.FileInfo, error)
		Open func(name string) (io.ReadCloser, error)
	}{
		Stat: os.Stat,
		Open: func(name string) (io.ReadCloser, error) { return os.Open(name) },
	}
	mimeTypes = map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".bmp":  "image/bmp",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
		".tif":  "image/tiff",
		".tiff": "image/tiff",
		".ico":  "image/x-icon",
		".pdf":  "application/pdf",
	}
)

func Upload(batchUploadRequest BatchUploadRequest) ([]Reference, error) {
	type response struct {
		resourceRef Reference
		err         error
	}

	responses, uploadWorkers := make(chan response, 1), sync.WaitGroup{}
	ctx, ctxCancelFunc := context.WithCancel(context.Background())

	uploadWorkers.Add(len(batchUploadRequest.Files)) // add an extra value to the counter for the routine that dequeues the responses

	for _, f := range batchUploadRequest.Files {
		go func(f string) { // upload files concurrently
			defer func() {
				uploadWorkers.Done()

				if err := recover(); err != nil {
					responses <- response{err: fmt.Errorf("panic in file upload func for file '%v'. %v", f, err)}
				}
			}()

			log.DebugPrintf("started file upload worker", "type", "batch_upload_request", "file", f)

			select {
			case <-ctx.Done():
				log.DebugPrintf("file upload context cancelled", "type", "batch_upload_request", "file", f)
				return
			default:
			}

			resourceRef, err := batchUploadRequest.UploadFunc(UploadRequest{
				URL:        batchUploadRequest.URL,
				Credential: batchUploadRequest.Credential,
				File:       f,
			})

			if err != nil {
				responses <- response{err: fmt.Errorf("unable to upload file '%v' via file storage api. %w", f, err)}
				return
			}

			responses <- response{resourceRef: resourceRef}
			log.DebugPrintf("stopped file upload worker", "type", "batch_upload_request", "file", f)
		}(f)
	}

	var (
		err          error
		resourceRefs = make([]Reference, 0, len(batchUploadRequest.Files))
	)

	responseWorkerDone := make(chan struct{}, 1)

	go func() {
		defer func() { responseWorkerDone <- struct{}{} }()

		log.DebugPrintf("started file upload response processing worker", "type", "batch_upload_response")

		for response := range responses {
			log.DebugPrintf("processing file upload response", "type", "batch_upload_response", "file", response.resourceRef.Label, "err", response.err)

			if err != nil { // once an error has occurred just drain the channel
				continue
			}

			if response.err != nil { // an error occurred...
				log.DebugPrintf("file upload response error. terminating workers", "type", "batch_upload_response", "err", response.err)
				err = response.err // ... and capture the error
				ctxCancelFunc()    // ... cancel any outstanding uploads ...
				continue
			}

			resourceRefs = append(resourceRefs, response.resourceRef)
		}

		log.DebugPrintf("stopped file upload response processing worker", "type", "batch_upload_response")
	}()

	uploadWorkers.Wait() // wait until all files have been uploaded
	close(responses)     // close the channel to terminate the response processing routine
	<-responseWorkerDone // wait until the response processing routine ends

	if err != nil {
		return nil, fmt.Errorf("unable to upload files to storage provider. %w", err)
	}

	return resourceRefs, nil
}

func FileInfo(file string) (os.FileInfo, string, error) {
	contentType := mimeTypes[filepath.Ext(file)]

	if contentType == "" {
		contentType = "text/plain"
	}

	fileInfo, err := FileIO.Stat(file)

	if err != nil {
		return nil, "", fmt.Errorf("invalid filepath. '%v' file does not exist. %w", file, err)
	}

	return fileInfo, contentType, nil
}
