package gemini

import (
	"encoding/json"
	"fmt"
)

type (
	CommandResult struct {
		Executed bool   `json:"executed"`
		Code     int    `json:"code"`
		Stderr   string `json:"stderr"`
		Stdout   string `json:"stdout"`
	}
	FilesRequestResult struct {
		Attached bool `json:"attached"`
	}
	CommandRequest struct {
		Text string `json:"text"`
	}
	FilesRequest struct {
		Files []string `json:"files"`
	}
)

type (
	commandExecutionTool struct{}
	googleSearchTool     struct{}
)

func (c googleSearchTool) marshalJSON() json.RawMessage {
	return json.RawMessage(`{
      "googleSearch": {}
    }`)
}

func (c commandExecutionTool) ExecCmdFunctionName() string {
	return "execute-command"
}

func (c commandExecutionTool) RequestFilesFunctionName() string {
	return "request-files"
}

func (c commandExecutionTool) marshalJSON() json.RawMessage {
	cmdExecFunctionDesc := fmt.Sprintf("executes a command on the user's machine. it runs as the user and you can consider it equivalent to you having access to their terminal. this command is primarily to be used to perform local "+
		"operations, such as querying or interacting with the file system or a local git repo. however, you may also use curl, wget and similar commands, if the user has explicitly asked you to do so, or it is implicit in the nature of their "+
		"request, such as a file download, api or web access. "+
		""+
		"when you choose to execute a command using this function you must provide a statement in your response alongside the command text that states the command you are executing. if it is "+
		"not clear from the user's original prompt, also include in this text why you are executing the command. this should be a very concise statement, without pleasantries like 'OK', and in the present "+
		"tense. for example: "+
		"'Executing git diff HEAD~1 to ascertain the changes made since your last git commit...'. the trailing '...' is important as it implies activity is occurring. "+
		""+
		"when the command spans more than one statement, join these lines with a ';', do not use new lines. similarly, never add new lines for formatting, present it all as a single line. such formatting is not required as "+
		"it can cause encoding errors with regexes and other escaped characters in the command, makes logs less readable as they need escaping and they are removed by the client before the user sees them anyway. "+
		""+
		"consider when structuring commands that it may result in better data transmission effificiencies have commands write to output local temporary files. you can then issue a follow up %v function call to upload those files for "+
		"for processing. when this completed. you can then issue a final command to remove those temporary files, however, ensure you only delete those temporary files you created, and no others. "+
		""+
		"when interpreting the result of the command consider that a successful command will always have a return code of '0'. the return code is passed in the returnCode field ('function_response.response.returnCode') and "+
		"the content of that field is the only value you need to ascertain success. if output from the command is expected, this will be found in the stdout field (`function_response.response.stdout`); this data represents the "+
		"data written to stdout by the command when it was executed on the user's machine. it is important to note that many commands do not write any data to stdout at all. so do not interpret an empty stdout "+
		"value as a failure. for example commands to write or delete files will not return any data, only a return code of 0, which, as explained earlier, you will interpret as success. "+
		""+
		"an unsuccessful command will always have a non-zero return code and it will likely also provide text explaining that error in the 'function_response.response.stderr' field. if the user declined to execute the command, "+
		"or cancelled it before it completed, the non-zero return code will be 125. "+
		""+
		"you must never repeatedly execute the same command. regardless of exit code. if you do not get the expected return code or stdout content. instead, terminate the conversation at that point and provide a response that summarises whatever "+
		"progress you made up to that point and then explains what it was about the last response that you considered incorrect. "+
		""+
		"when you provide your final response to the user based on the output or return code of the command execution, do one of the following. if the return code is 0 (success) and there is data in stdout, "+
		"respond with that output only, and absolutely nothing else, that way it can be piped into another command. if the return code is 0 (success) but there is no data in stdout, respond only with the word 'OK', "+
		"and absolutely nothing else, that way it can be piped into another command. finally, if the return code is not 0 (error) respond only with the word 'Error' followed by any data in stderr. "+
		""+
		"in the event the user's instructions require you to terminate the process with a particular exit code, the exact command required for that is simply 'exit {code}'. do not try to kill other processes or the terminal, "+
		"just exit the current one using that command", c.RequestFilesFunctionName())
	uploadFilesFunctionDesc := fmt.Sprintf("provides the content of all the files in the user's file system that are listed in the 'files' arguments. this is to be used in support of the '%v' as a more efficient "+
		"alternative to accessing file contents by directly executing a shell command. use this function instead of executing 'cat file', for example. you can also use it upload data you have generated yourself "+
		"more efficiently. for example if the user requests a command be executed, you could redirect the output to a file, then request that file using this function, then delete it after", c.ExecCmdFunctionName())
	return json.RawMessage(fmt.Sprintf(`{
      "functionDeclarations": [
		{ 
			"name": "%v",
			"description": "%v", 
			"parameters": { 
				"type": "object", 
				"properties": { 
					"text":  { "type": "string", "description": "the complete text of the command to be executed in the shell, for example, if asked to to get the number of files in the current directory; 'ls -l | wc -l' may be specified" } 
				} 
			} 
		},
		{ 
			"name": "%v",
			"description": "%v", 
			"parameters": { 
				"type": "object", 
				"properties": { 
					"files":  { "type": "array", "items": { "type": "string" }, "description": "the files to upload. each file should specified using its relative path" } 
				} 
			} 
		}
	]}`, c.ExecCmdFunctionName(), cmdExecFunctionDesc, c.RequestFilesFunctionName(), uploadFilesFunctionDesc))
}

func (c CommandResult) marshalJSON() json.RawMessage {
	j, _ := json.Marshal(map[string]any{
		"name": (commandExecutionTool{}).ExecCmdFunctionName(),
		"response": map[string]any{
			"returnCode": c.Code,
			"stdErr":     c.Stderr,
			"stdOut":     c.Stdout,
		},
	})

	return j
}

func (c FilesRequestResult) marshalJSON() json.RawMessage {
	j, _ := json.Marshal(map[string]any{
		"name": (commandExecutionTool{}).RequestFilesFunctionName(),
		"response": map[string]any{
			"attached": c.Attached,
		},
	})

	return j
}

func (c CommandRequest) marshalJSON() json.RawMessage {
	j, _ := json.Marshal(c)

	return j
}

func (f FilesRequest) marshalJSON() json.RawMessage {
	j, _ := json.Marshal(f)

	return j
}
