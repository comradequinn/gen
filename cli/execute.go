package cli

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/comradequinn/gen/gemini"
	"github.com/comradequinn/gen/log"
)

func execute(output gemini.Output, cfg gemini.Config, scriptMode bool) (gemini.CommandResult, error) {
	result := gemini.CommandResult{
		Executed: true,
	}

	if output.Text != "" && !scriptMode {
		WriteInfo("%v", output.Text)
	}

	if output.CommandRequest.Text == "" {
		return result, fmt.Errorf("command text is empty")
	}

	log.DebugPrintf("executing command locally", "type", "cmd_executing", "text", output.CommandRequest.Text)

	if cfg.CommandApproval {
		Write("approval is required for the execution of the following script:\n\n")
		WriteInfo(output.CommandRequest.Text + "\n")
		Write("enter 'y' to approve the execution. enter any other value to deny: ")

		input, _, _ := bufio.NewReader(Reader).ReadRune()

		if strings.ToLower(string(input)) != "y" {
			log.DebugPrintf("command execution declined by user", "type", "cmd_execution_declined", "text", output.CommandRequest.Text)
			result.Code = 125
			return result, nil
		}
		Write("")
	}

	cmd := exec.Command("bash", "-c", output.CommandRequest.Text)

	cmd.Stdout = &bytes.Buffer{}
	cmd.Stderr = &bytes.Buffer{}

	err := cmd.Run()

	result.Stdout = (cmd.Stdout.(*bytes.Buffer)).String()
	result.Stderr = (cmd.Stderr.(*bytes.Buffer)).String()

	if err != nil {
		result.Code = 127 // unrecognised or unexecutable command

		if exitErr, ok := err.(*exec.ExitError); ok {
			result.Code = exitErr.ExitCode()
		}
	}

	log.DebugPrintf("executed command locally", "type", "cmd_executed", "text", output.CommandRequest.Text, "code", result.Code, "stdout", string(result.Stdout), "stderr", string(result.Stderr))

	return result, nil
}
