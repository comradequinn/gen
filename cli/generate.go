package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/comradequinn/gen/gemini"
	"github.com/comradequinn/gen/log"
	"github.com/comradequinn/gen/session"
)

func Generate(cfg gemini.Config, args Args, scriptMode bool, promptText, schema string, files []string) {
	var err error

	generate := func(prompt gemini.Prompt) gemini.Transaction {
		stopSpinner := func() {}
		if !scriptMode {
			stopSpinner = spin()
		}

		prompt.History, err = session.Read(*args.AppDir)
		log.FatalfIf(err != nil, "unable to read history. %v", err)

		transaction, err := gemini.Generate(cfg, prompt)

		log.FatalfIf(err != nil, "error with gemini api. %v", err)

		log.FatalfIf(session.Write(*args.AppDir, transaction) != nil, "unable to update session. %v", err)

		stopSpinner()

		if *args.Stats {
			enc := json.NewEncoder(os.Stderr)
			enc.SetIndent("", "  ")
			_ = enc.Encode(map[string]map[string]string{
				"stats": {
					"systemPromptBytes": fmt.Sprintf("%v", len(*args.SystemPrompt)),
					"promptBytes":       fmt.Sprintf("%v", len(prompt.Text)),
					"responseBytes":     fmt.Sprintf("%v", len(transaction.Output.Text)),
					"tokens":            fmt.Sprintf("%v", transaction.Tokens),
					"files":             fmt.Sprintf("%v", len(transaction.Input.FileReferences)),
				},
			})
		}

		return transaction
	}

	prompt := gemini.Prompt{
		Text:      promptText,
		Files:     files,
		InputType: gemini.InputTypeUser,
		Schema:    gemini.JSONSchema(schema),
	}

	transaction := generate(prompt)

	for transaction.Output.IsFunction() {
		prompt := gemini.Prompt{
			InputType: gemini.InputTypeCommand,
		}

		switch {
		case transaction.Output.IsCommandRequest():
			if prompt.CommandResult, err = execute(transaction.Output, cfg, scriptMode); err != nil {
				log.FatalfIf(err != nil, "error executing command '%v' on behalf of gemini. %v", transaction.Output.CommandRequest.Text, err)
			}
			if prompt.CommandResult.Code != 0 && scriptMode {
				log.DebugPrintf(fmt.Sprintf("terminating with non-zero exit code as script/quiet mode was enabled when a command executed on behalf of gemini signalled the exit code %v", prompt.CommandResult.Code))
				os.Exit(prompt.CommandResult.Code)
			}
		case transaction.Output.IsFilesRequest():
			prompt.Files = transaction.Output.FilesRequest.Files
			prompt.FilesRequestResult = gemini.FilesRequestResult{Attached: true}
		}

		transaction = generate(prompt)
	}

	Write("%v\n", transaction.Output.Text)
}
