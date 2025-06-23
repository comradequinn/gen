package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/comradequinn/gen/gemini"
	"github.com/comradequinn/gen/log"
	"golang.org/x/sync/errgroup"
)

func writeFiles(request gemini.WriteRequest, scriptMode bool) (gemini.WriteResult, error) {
	g, ctx := errgroup.WithContext(context.Background())

	for _, f := range request.Files {
		if !scriptMode {
			WriteInfo("writing %v bytes to file '%v' ....", len(f.Data), f.Name)
		}

		g.Go(func() error {
			select {
			case <-ctx.Done():
				log.DebugPrintf("local file write cancelled due to related writes already having errored", "type", "file_write_cancelled", "file", f.Name, "len", len(f.Data))
				return ctx.Err()
			default:
			}

			log.DebugPrintf("writing file locally", "type", "writing_file", "file", f.Name, "len", len(f.Data))

			if err := os.MkdirAll(filepath.Dir(f.Name), 0755); err != nil {
				return fmt.Errorf("unable to verify or create directory for write-request for '%v'. %w", f.Name, err)
			}

			file, err := os.Create(f.Name)
			if err != nil {
				return fmt.Errorf("unable to create file for write-request for '%v'. %w", f.Name, err)
			}
			defer file.Close()

			if _, err := file.WriteString(f.Data); err != nil {
				return fmt.Errorf("unable to write data to file '%v' for write-request. %w", f.Name, err)
			}

			log.DebugPrintf("file written locally", "type", "file_written", "file", f.Name, "len", len(f.Data))

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return gemini.WriteResult{}, err
	}

	return gemini.WriteResult{Written: true}, nil
}
