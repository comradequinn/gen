package cli

import (
	"github.com/comradequinn/gen/gemini"
	"github.com/comradequinn/gen/log"
)

func readFiles(request gemini.ReadRequest, quiet bool) ([]string, gemini.ReadResult) {
	for _, f := range request.FilePaths {
		log.DebugPrintf("local file requested", "type", "file_request", "file", f)

		if !quiet {
			WriteInfo("reading file '%v'...", f)
		}
	}

	return request.FilePaths, gemini.ReadResult{FilesAttached: true}
}
