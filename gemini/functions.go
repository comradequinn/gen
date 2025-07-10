package gemini

import (
	"encoding/json"
	"fmt"
)

type (
	ExecuteResult struct {
		Executed bool   `json:"executed"`
		Code     int    `json:"code"`
		Stderr   string `json:"stderr"`
		Stdout   string `json:"stdout"`
	}
	ReadResult struct {
		FilesAttached bool `json:"filesAttached"`
	}
	ExecuteRequest struct {
		Text string `json:"text"`
	}
	ReadRequest struct {
		FilePaths []string `json:"filePaths"`
	}
	WriteRequest struct {
		Files []File `json:"files"`
	}
	File struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}
	WriteResult struct {
		Written bool `json:"written"`
	}
)

type (
	executeTool      struct{}
	googleSearchTool struct{}
)

func (c googleSearchTool) marshalJSON() json.RawMessage {
	return json.RawMessage(`{
      "googleSearch": {}
    }`)
}

func (c executeTool) ExecuteFunctionName() string {
	return "execute"
}

func (c executeTool) ReadFunctionName() string {
	return "read"
}

func (c executeTool) WriteFunctionName() string {
	return "write"
}

func (c executeTool) marshalJSON() json.RawMessage {
	executeFunctionDesc := fmt.Sprintf("executes a command on the user's machine. it runs as the user and you can consider it equivalent to you having access to their terminal. this command is primarily to be used to perform local "+
		"operations, such as querying or interacting with the file system or a local git repo. however, you may also use curl, wget and similar commands, if the user has explicitly asked you to do so, or it is implicit in the nature of their "+
		"request, such as a file download, api or web access. "+
		""+
		"when you choose to execute a command using this function do not provide any textual information to accompany it, as this will not be shown to the user.  "+
		""+
		"when the command spans more than one statement, join these lines with a ';', do not use new lines. similarly, never add new lines for formatting, present it all as a single line. such formatting is not required as "+
		"it can cause encoding errors with regexes and other escaped characters in the command, makes logs less readable as they need escaping and they are removed by the client before the user sees them anyway. "+
		""+
		"consider when structuring commands that it may result in better data transmission efficiencies to have commands write to output local temporary files. you can then issue a follow up %v function call to upload those files for "+
		"for processing. when you do create a file purely for the purposes of uploading data, you should issue a follow up command to remove those files to avoid leaving unexpected files on the users disk. "+
		""+
		"do not include 'sudo' in any of your commands. just attempt the command without it. if the user does not have sufficient permissions, suggest they re-run you as sudo adn retry. "+
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
		"just exit the current one using that command", c.ReadFunctionName())
	readFunctionDesc := fmt.Sprintf("provides the content of all the files in the user's file system that are listed in the 'filePaths' arguments. you may use this to access local files that the user has referred to in their prompt in "+
		"order to provide you with any required context. for example, if a user refers to the 'my data.txt' file or 'the Dockerfile', you can use this to view the contents of those files and help you process their request. "+
		"this is also to be used in support of the '%v' function as a more efficient alternative to accessing file contents by directly executing a command. use this function instead of "+
		"executing 'cat file', for example. you can also use it upload data you have generated yourself more efficiently. for example if the user requests a command be executed, you could redirect the output to a file, then request that "+
		"file using this function. ", c.ExecuteFunctionName())
	writeFunctionDesc := fmt.Sprintf("writes files to the users files system as specified in the files argument. this is to be used in support of the '%v' function as a more efficient "+
		"and effective alternative to writing or modifying file contents by directly executing commands. for example, you could use this function instead of executing the command 'echo data > file.txt' or to avoid defining commands "+
		"with complex transforms, using sed, grep and similar, to apply your required edits to files. Instead, just use this function to state what the exact contents of files should be. You can still use commands if that approach would be "+
		"simpler, but for large files or complex edits, this function may be preferable", c.ExecuteFunctionName())
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
					"filePaths":  { "type": "array", "items": { "type": "string" }, "description": "the files to upload from the user's filesystem. each file should specified using its relative path" } 
				} 
			} 
		},
		{ 
			"name": "%v",
			"description": "%v", 
			"parameters": { 
				"type": "object", 
				"properties": { 
					"files":  { "type": "array", "items": 
						{ 
							"type": "object", 
						 	"properties": { 
								"name": { 
									"type": "string", 
									"description": "the path of the file to write. for example '.data/myfile.txt' or './myfile.txt'" 
								},
								"data": { 
									"type": "string", 
									"description": "the full content of the file" 
								}
							} 
						}, "description": "the files to write to the user's file system" 
					} 
				} 
			} 
		}
	]}`, c.ExecuteFunctionName(), executeFunctionDesc, c.ReadFunctionName(), readFunctionDesc, c.WriteFunctionName(), writeFunctionDesc))
}

func (c ExecuteResult) marshalJSON() json.RawMessage {
	j, _ := json.Marshal(map[string]any{
		"name": (executeTool{}).ExecuteFunctionName(),
		"response": map[string]any{
			"returnCode": c.Code,
			"stdErr":     c.Stderr,
			"stdOut":     c.Stdout,
		},
	})

	return j
}

func (r ReadResult) marshalJSON() json.RawMessage {
	j, _ := json.Marshal(map[string]any{
		"name": (executeTool{}).ReadFunctionName(),
		"response": map[string]any{
			"attached": r.FilesAttached,
		},
	})

	return j
}

func (r WriteResult) marshalJSON() json.RawMessage {
	j, _ := json.Marshal(map[string]any{
		"name": (executeTool{}).WriteFunctionName(),
		"response": map[string]any{
			"written": r.Written,
		},
	})

	return j
}

func (c ExecuteRequest) marshalJSON() json.RawMessage {
	j, _ := json.Marshal(c)
	return j
}

func (r ReadRequest) marshalJSON() json.RawMessage {
	j, _ := json.Marshal(r)
	return j
}

func (w WriteRequest) marshalJSON() json.RawMessage {
	j, _ := json.Marshal(w)
	return j
}
