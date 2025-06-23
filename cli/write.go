package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/comradequinn/gen/gemini"
	"github.com/comradequinn/gen/log"
)

func writeFiles(request gemini.WriteRequest, scriptMode bool) (gemini.WriteResult, error) {
	for _, f := range request.Files {
		if !scriptMode {
			WriteInfo("writing %v bytes to file '%v' ....", len(f.Data), f.Name)
		}

		log.DebugPrintf("writing file locally", "type", "writing_file", "file", f.Name, "len", len(f.Data))

		if err := os.MkdirAll(filepath.Dir(f.Name), 0755); err != nil {
			return gemini.WriteResult{}, fmt.Errorf("unable to verify or create directory for write-request for '%v'. %w", f.Name, err)
		}

		file, err := os.Create(f.Name)
		if err != nil {
			return gemini.WriteResult{}, fmt.Errorf("unable to create file for write-request for '%v'. %w", f.Name, err)
		}
		defer file.Close()

		if _, err := file.WriteString(f.Data); err != nil {
			return gemini.WriteResult{}, fmt.Errorf("unable to write data to file '%v' for write-request. %w", f.Name, err)
		}

		log.DebugPrintf("file written locally", "type", "file_written", "file", f.Name, "len", len(f.Data))
	}

	return gemini.WriteResult{Written: true}, nil
}
