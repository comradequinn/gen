package gemini

import "github.com/comradequinn/gen/gemini/internal/schema"

func addHistory(transactions []Transaction) []schema.Content {
	contents := make([]schema.Content, 0, len(transactions)+1)

	for _, transaction := range transactions {
		content := schema.Content{
			Role: RoleUser,
		}

		switch {
		case transaction.Input.IsExecuteResult():
			content.Parts = append(content.Parts, schema.Part{FunctionResponse: transaction.Input.ExecuteResult.marshalJSON()})
		case transaction.Input.IsReadResult():
			content.Parts = append(content.Parts, schema.Part{FunctionResponse: transaction.Input.ReadResult.marshalJSON()})
		case transaction.Input.IsWriteResult():
			content.Parts = append(content.Parts, schema.Part{FunctionResponse: transaction.Input.WriteResult.marshalJSON()})
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

		if transaction.Output.IsExecuteRequest() {
			content.Parts = append(content.Parts, schema.Part{FunctionCall: schema.FunctionCall{
				Name: (executeTool{}).ExecuteFunctionName(),
				Args: transaction.Output.ExecuteRequest.marshalJSON(),
			}})
		}

		if transaction.Output.IsReadRequest() {
			content.Parts = append(content.Parts, schema.Part{FunctionCall: schema.FunctionCall{
				Name: (executeTool{}).ReadFunctionName(),
				Args: transaction.Output.ReadRequest.marshalJSON(),
			}})
		}

		if transaction.Output.IsWriteRequest() {
			content.Parts = append(content.Parts, schema.Part{FunctionCall: schema.FunctionCall{
				Name: (executeTool{}).WriteFunctionName(),
				Args: transaction.Output.WriteRequest.marshalJSON(),
			}})
		}

		contents = append(contents, content)
	}

	return contents
}
