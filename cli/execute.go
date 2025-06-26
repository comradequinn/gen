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

func execute(request gemini.ExecuteRequest, cfg gemini.Config, quiet bool) (gemini.ExecuteResult, error) {
	result := gemini.ExecuteResult{
		Executed: true,
	}

	if request.Text == "" {
		return result, fmt.Errorf("command text is empty")
	}

	if !quiet {
		WriteInfo("executing... [%v]", request.Text)
	}

	log.DebugPrintf("executing command locally", "type", "cmd_executing", "text", request.Text)

	if cfg.ExecutionApproval {
		Write("approval is required for the execution of the following:\n\n")
		WriteInfo(request.Text + "\n")
		Write("enter 'y' to approve the execution. enter any other value to deny: ")

		input, _, _ := bufio.NewReader(Reader).ReadRune()

		if strings.ToLower(string(input)) != "y" {
			log.DebugPrintf("command execution declined by user", "type", "cmd_execution_declined", "text", request.Text)
			result.Code = 125
			return result, nil
		}
		Write("")
	}

	cmd := exec.Command("bash", "-c", request.Text)

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

	log.DebugPrintf("executed command locally", "type", "cmd_executed", "text", request.Text, "code", result.Code, "stdout", string(result.Stdout), "stderr", string(result.Stderr))

	return result, nil
}
