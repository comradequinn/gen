package gemini

import "github.com/comradequinn/gen/gemini/internal/schema"

func historyFrom(prompt Prompt) []schema.Content {
	contents := make([]schema.Content, 0, len(prompt.History)+1)

	for _, transaction := range prompt.History {
		content := schema.Content{
			Role: RoleUser,
		}

		switch {
		case transaction.Input.IsCommandResult():
			content.Parts = append(content.Parts, schema.Part{FunctionResponse: transaction.Input.CommandResult.marshalJSON()})
		case transaction.Input.IsFilesRequestResult():
			content.Parts = append(content.Parts, schema.Part{FunctionResponse: transaction.Input.FilesRequestResult.marshalJSON()})
		default:
			if transaction.Input.Text != "" {
				content.Parts = append(content.Parts, schema.Part{Text: transaction.Input.Text})
			}

			if len(transaction.Input.FileReferences) > 0 {
				for _, fileReference := range transaction.Input.FileReferences {
					content.Parts = append(content.Parts, schema.Part{
						File: &schema.FileData{URI: fileReference.URI, MIMEType: fileReference.MIMEType},
					})
				}
			}
		}

		contents = append(contents, content)

		content = schema.Content{
			Role: RoleModel,
		}

		if transaction.Output.Text != "" {
			content.Parts = append(content.Parts, schema.Part{Text: transaction.Output.Text})
		}

		if transaction.Output.IsCommandRequest() {
			content.Parts = append(content.Parts, schema.Part{FunctionCall: schema.FunctionCall{
				Name: (commandExecutionTool{}).ExecCmdFunctionName(),
				Args: transaction.Output.CommandRequest.marshalJSON(),
			}})
		}

		if transaction.Output.IsFilesRequest() {
			content.Parts = append(content.Parts, schema.Part{FunctionCall: schema.FunctionCall{
				Name: (commandExecutionTool{}).RequestFilesFunctionName(),
				Args: transaction.Output.FilesRequest.marshalJSON(),
			}})
		}

		contents = append(contents, content)
	}

	return contents
}
