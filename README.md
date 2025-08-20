# gen

Named `gen` (from `generate`), `gen` is an agentic, command-line `llm` interface built on Google's `Gemini` models.

Using `gen` greatly simplifies integrating LLM capabilities into CI pipelines, scripts or other automation.

For terminal users, `gen` acts as a fully-featured conversational, agentic assistant.

## Features

Using `gen` provides the following features:

* Conversational, command-line chat with support for multiple conversation sessions

* Agentic capabilities, ask `gen` to `do` a task for you rather than explain how

* Define `structured responses` using custom `schemas`

* Fully scriptable and ideal for use in automation and CI pipelines

* Include text, code, image and pdf files in prompts; either explicitly or let `gen` discover and upload what it needs to complete the task
 * Model configuration; specify general and use-case based `system-prompt` content, custom `models`, `temperature` and `top-p` to fine-tune output

## Quick Start

These examples show typical sequences for the two forms of `gen` usage: interactive and scripted. For those familiar with the command line, scripting and LLMs, they are likely enough for you to become productive with `gen`.

Installation instructions can be found [here](./INSTALL.md).

### Conversational Basics

The following examples show the fundamentals of conversational usage of `gen` within a user's terminal

```bash
# asking a question initiates a new conversation
gen "what is the latest version of go?"
# >> The latest stable version of Go is.... (response truncated for brevity)

# ask a follow up question (by passing the -c/--continue flag)
gen -c "what were the major amendments in that release?"
# >> the release introduced several significant amendments across its toolchain, runtime and... (response truncated for brevity)

# stash the existing conversational context and start a new session (by omitting the -c/--continue flag)
gen "what is the weather like in london tomorrow?"
# >> In London tomorrow it will be grey and wet... (response truncated for brevity)

# show current and active conversation sessions, the asterix indicates the active session (-l is the shortform of --list)
gen -l
   #1 (April 26 2025): what is the latest version of go?
*  #2 (April 26 2025): what is the weather like in london tomorrow?

# switch the active session back to the earlier topic (-r is the shortform of --restore)
gen -r 1

# ask a follow up question relying on context from the restored session (by passing the -c/--continue flag)
gen -c "was this a major or minor release?"
# >> It was a minor release. Go follows a versioning scheme that.... (response truncated for brevity)

# explicitly attach a file to the prompt and ask a question related to its contents (-f is the  shortform of --files)
# use the pro model for more complex tasks (by passing the -pro flag)
gen -f -pro ./main.go "perform a code review on this file"
# >> This is a solid, well-written `main.go` file. It demonstrates good... (response truncated for brevity)
```

### Agentic Actions

When `exec` mode is enabled, `gen` can perform actions on the host machine, such as reading and writing files and executing commands.

```bash
# use exec mode to allow gen to directly execute commands to complete tasks (-x is the shortform of --exec)...
# ...this is the same request as above but without attaching the file explicitly
gen -x "summarise the code in main.go"
# >> reading file 'main.go'...
# >> This is a Go program, named 'gen'. It functions as a command-line interface for interacting.... (response truncated for brevity)

# directly execute and interpret terminal commands to perform tasks
gen -x "list all .go files in my current directory"
# >> executing... [ls -l *.go]
# >> main.go

# execute a follow up task using the previous task for context (by passing the -c/--continue flag)
gen -c -x "copy the files to a directory named 'backup'"
# >> executing... [mkdir -p backup; cp *.go backup/]
# >> OK

# access and parse external resources
gen -x "write the contents returned from github's home page to a file named github.html"
# >> executing... [curl https://github.com > github.html]
# >> OK

# interact with the host system
gen -x "get the pids, names and cpu and mem usage of the top 5 processes running by cpu. format this in markdown as a table"
# >> executing... [ps -eo pid,cmd,%cpu,%mem --sort=-%cpu | head -n 6]
# >>
# >> | PID | CMD | %CPU | %MEM |
# >> |---|---|---|---|
# >> | 1331 | /home/me/process/job1 | 6.2 | 5.6 |
# >> | 6911 | /home/me/process/code | 1.4 | 0.6 |
# >> | 519 | /home/me/process/job2 | 1.1 | 0.7 |
# >> | 1420 | /home/me/process/task1 | 0.6 | 2.7 |
# >> | 2058 | /home/me/process/job3 | 0.5 | 0.4 |
```

### Scripting

By defining [structured responses](#structured-responses) and controlling output, `gen` can be used to bring LLM capabilities to scripts, ci or other automation.

```bash
# suppress non-result output with (-q is the shortform of --quiet)
# and redirect the response to file
gen -q "pick a colour of the rainbow" > colour.txt
# >> file: colour.txt
# >> blue

# instruct gen to return different exit codes based on specified conditions
gen -q -x "terminate this process with an exit code of 1 if it is monday, 2 if is tuesday, or 3 otherwise"
## > OK
echo "exit code: $?"
## > 3 (assuming it is not monday or tuesday)

# provide a custom json schema for the response (-s is the shortform of --schema)
# - use GSL (gen schema language) to concisely define an array of objects each with a string and boolean field 
# - the openAPI spec can also be used for more complex requirements
# - use the response in a script; in this case it is simply piped into jq for formatting purposes  
gen -q -s '[]colour:string|primary:boolean:true if primary' "list all colours of the rainbow" | jq
## > [
## >   {
## >     "colour": "Red",
## >     "primary": true
## >   },
## >   {
## >     "colour": "Orange",
## >     "primary": false
## >   },
## >   {
## >     "colour": "Yellow",
## >     "primary": false
## >   },
## >   {
## >     "colour": "Green",
## >     "primary": true
## >   },
## >   (response truncated for brevity)
```

## Installation

Installation instructions are available [here](./INSTALL.md).

## Usage

### Prompting

To chat with `gen`, execute it with a prompt to start a new conversation session. The result of the prompt will be displayed in response, as shown below.

```bash
gen "how do I list all files in my current directory?"
# >> To list all files in your current directory, you can use the following command in your terminal:
# >> ls -a
# >> This command will display all files, including hidden files (files starting with a dot).
```

To ask a follow up question, run `gen` again with the required prompt and specify `-c` or `--continue`. The conversation context from the previous prompt will then be maintained.

```bash
gen -c "I need timestamps in the output"
# >> To include timestamps in the output of the `ls` command, you can use the `-l` option along with the `--full-time` or `--time-style` options
```

This conversational context will be maintained for as long you pass  `-c` or `--continue`. Start a new session by omitting it. When a new session is started, the existing session is `stashed`. As shown below

```bash
gen "what was my last question?"
# >> I have no memory of past conversations. Therefore, I don't know what your last question was.
```

To view your previously `stashed` sessions, run `gen --list` (or `-l`). The sessions will be displayed in date order and include a snippet of the opening text of the prompt for ease of identification. The active session is also included in the output and prefixed with an asterix, in this case record `2`.

```bash
gen -l
  #1 (April 15 2025): 'how do i list all files in my current directory?'
* #2 (April 15 2025): 'what was my last question?'
```

To restore a previous session, allowing you to continue that conversation as it was where you left off, run `gen --restore #id` (or `-r`) where `#id` is the `#ID` in the `gen --list` output. For example

```bash
gen --r 1
```

Running `gen -l` again will now show the below; note how the asterisk is now positioned at record `1`

```bash
gen -l
* #1 (April 15 2025): 'how do i list all files in my current directory?'
 #2 (April 15 2025): 'what was my last question?'
```

Asking the prompt from earlier for which `gen` had no context, along with the `-c` or `--continue` flag, will now return the below, as that context has been restored.

```bash
gen -c "what was my last question?"`
# >> Your last question was: "I need timestamps in the output".
```

To delete a single session, run `gen --delete #id` (or `-d #id`) where `#id` is the `#ID` in the `gen --list` output. To delete all sessions, run `gen --delete-all`

### Agentic Actions

To run `gen` in `exec` mode, pass the `--exec` (or `-x`) flag.

When running in `exec` mode, `gen` behaves `agentically`. It is able to execute commands as your user in order to perform tasks on your behalf. These tasks can effectively be anything that you could undertake yourself and can also contain multiple steps. The exception to this is long running programs. Asking `gen` to capture all tcp traffic to a host will work, but if you do not also specify some form of exit condition, the command will run indefinitely and the result will never be collected and forwarded to `gemini` in order for it to respond; `gen` will just appear to hang.

The `--exec` flag is scoped to each individual prompt, so agentic capabilities can be variably enabled or disabled on individual prompts within the same conversation.

For security purposes, it is impossible for `gen` to execute commands without the `--exec` (or `-x`) flag, and conversely, take extra caution with your prompts when `exec` mode is enabled. If you would prefer to approve each command that `gen` requests before it is executed, pass the `--approve` (or `-k`) flag along with `--exec`.

> Note that `grounding` will be implicitly disabled when running in `exec` mode. This is a current stipulation of the `Gemini API`, not `gen` itself. However, this can easily be mitigated by running an initial prompt with `exec` enabled to take whatever agentic actions are needed, and then running subsequent prompts without `exec` mode enabled. The context of the `exec` enabled prompts will still be present in the later `non-exec` prompts, but grounding will be available to enhance the capabilities of the model's interactions with that data.

An example is shown below of using agentic mode in a conversation.

```bash
# ask for support with a task, as --exec is not specified, gen will explain how to do something, rather than simply doing it
gen "how would I list all files in this directory, excluding any git files, that were modified in the last day?"
# >> To list all files in the current directory, excluding any git-related files, that were modified in the last day, use the following command:
# >>
# >> find . -type f -mtime -1 -not -path '*/.git/*' -not -name '.gitignore' -not -name '.gitmodules' -not -name '.gitattributes'

# switch to exec mode by passing --exec flag for the next prompt and use the previous context to infer the action gen is take
gen -c -x "ok, execute that for me and print the results"
# >>executing... [find . -type f -mtime -1 -not -path '*/.git/*' -not -name '.gitignore' -not -name '.gitmodules' -not -name '.gitattributes']
# >> ./gemini/config_test.go
# >> ./gemini/types.go
# >> ./gemini/gemini.go
# >> .... (response truncated for brevity)

# ask a further query, but this time without exec mode (but still with access to the data accessed with exec mode)
# this will enable grounding (though it would not be used this example)
gen -c "how many files was that in total?"
# >> There were 22 files in total.
```

As well as relatively specific tasks, such as those above, you can also give general instructions to `gen`, and have it figure the steps it needs to take and execute them. As illustrated below.

```bash
gen -x "can I use the code in this repo for commercial purposes?"
# >> executing... [ls -F]
# >> reading file 'LICENSE'...
# >> No, the GNU General Public License (GPL) version 3, under which this repository's code is licensed, does not permit.... (response truncated for brevity)
```
### Including Files

When `gen` is running in `exec` mode, it will dynamically identify any files it needs and upload them. As shown below.

```bash
gen -x "create a table of file names and a brief summary of the file's content for each .go file in this directory and all subdirectories"
# >> executing... [find . -name "*.go"]
# >> reading file './gemini/config_test.go'...
# >> reading file './gemini/types.go'...
# >> reading file './gemini/gemini.go'...
# >> ...(activity truncated for brevity)

# >> - ./gemini/config_test.go: Unit tests for the Gemini API client configuration.
# >> - ./gemini/types.go: Defines data structures for Gemini API interactions, including configurations, prompts, and transaction details.
# >> - ./gemini/gemini.go: Core logic for interacting with the Gemini API, handling content generation and function calls.
# >> - ./gemini/functions.go: Defines functions that the Gemini model can call, such as executing commands, reading, and writing files.
# >> ...(response truncated for brevity)
```

Alternatively, should you wish to control the exact files that are uploaded, or you wish to access files without enabling `exec` mode, you can attach them explicitly (and then, optionally, not enable `exec` mode at all).

To explicitly include files in your prompt, use the `--files` (or `-f`) parameter passing the path to the file to include. To include multiple files, separate them with a comma, Some examples are shown below.

```bash
# attach a single file
gen --files "holday-fun.png" "what's in this image?"
```

```bash
# attach multiple files, spaces are optional but can aid readability when listing many files explicitly. here we use the shortform of --files; -f
gen -f "some-code.go, somedir/some-more-code.go, yet-more-code.go" "summarise these files"
```

When attaching a large number of files or the contents of multiple directories, any `posix` compliant shell will support `command substitution` which can be used to simplify creating the `files` argument. An example is shown below of including all `*.go` files in the current workspace (that being the working directory and below).

```bash
# find all files in the current workspace and concatenate them into a single string
WORKSPACE="$(find . -name "*.go" | paste -s -d ",")"
# attach the files to the prompt
gen -f "$WORKSPACE" "create a table of file names and a very brief content summary for these files"
```

When run on the `gen` repo, the above will produce something similar to the below.

```bash
# >> | File | Summary |
# >> | --- | --- |
# >> | `main.go` | The main entry point for the `gen` command-line application. It handles argument parsing, command dispatching, and initialization. |
# >> | `cli/args.go` | Defines and parses all command-line flags and arguments for the application using Go's `flag` package. |
# >> | `cli/execute.go` | Contains the logic to execute shell commands requested by the Gemini model, including handling user approval for security. |

...(response truncated for brevity)
```

### Scripting

When using the output of `gen` in a script, it is advisable to suppress activity indicators and other interactive output using the `--quiet` flag (or `-q`). This ensures a consistent output stream containing only response data.

The simple example below uses redirection to write the response to a file.

```bash
gen --quiet "pick a colour of the rainbow" > colour.txt
```
This will result in a file similar to the below

```bash
# file: colour.txt
Blue
```

#### Agentic Actions in Scripts

Using `exec` mode in scripts is largely the same as using it conversationally. The key difference is that there is no user interaction. So `gen` cannot seek instruction on how to proceed after an error. 

Similarly, when commands executed by `gen` fail, a script would typically expect `gen` to terminate and pass that exit code back to it; which is indeed how it behaves. You can also instruct `gen` to exit with a specific return code in your prompt too.

The following example illustrates these concepts by scripting the automated code review of a given file. As this is a script, the longform of arguments is used for improved readability, though this is entirely optional.

```bash
gen --exec --quiet "perform a code review on the main.go file.
                    write a single integer quality result value to a new file named 'result.txt',
                    with 1 being 'excellent' and 5, 'poor'.
                    if the result is not 1, write a justification as to what needs improving
                    to a new file named 'feedback.txt'. finally, if the result was not 1, issue a
                    command to terminate the process with a return code of 2"

case $? in
   0)
       # take whatever action is required for a completed code review with a positive outcome...
       echo "code review completed. code is excellent. score $(cat result.txt)"
       ;;
   1)
       # take whatever action is required for code review that could not be completed due to an error executing the commands...
       echo "an error prevented the code review being undertaken"
       ;;
   2)
       # take whatever action is required for a completed code review with a negative outcome...
       echo "code review completed. code is poor. detailed feedback: $(cat feedback.txt)"
       ;;
   *)
       # default error case for any other exit codes
       echo "an unknown error occurred. exit code: $?"
       ;;
esac
```

When executed, the above will cause `gen` to silently perform all the tasks and the subsequent part of the script will then print an appropriate response based on `gen's` return code. In the prompt used, the `exit code` to indicate code review failure was set as `2`; this is optional and is purely to allow the two causes of failure to be distinguished in the example script.

#### Structured Responses

By default, `gen` will request responses structured as free-form text, which is a sensible format for conversational use. 

However, in many scenarios, particularly CI and scripting use-cases, it is preferable to have the output in a structured form so that it may be reliably interrogated and actions then taken based up on its content. 

To this end, `gen` allows you to specify a schema, using `--schema` (or `-s`), that will be used to structure the response.

There are two methods of specifying a schema, either by using `GSL` (`gen`'s `s`chema `l`anguage) or by providing a JSON based `OpenAPI schema object`.

In either case, note that `grounding` will be implicitly disabled when using a `schema`, this is a current stipulation of the `Gemini API`, not `gen` itself.

##### GSL (Gen's Schema Language)

`GSL` provides a quick, simple and readable method of defining basic response schemas. It allows the definition of an arbitrary number of `fields`, each with a `type` and an optional `description`. `GSL` can only be used to define non-hierarchical schemas, however this is often all that is needed for a substantial amount of structured response use-cases.

A basic schema definition in `GSL` format is shown below, it represents a single field response with no description

```bash
field-name:type # for example, 'result:integer'
```
A more complex definition showing multiple fields, each with descriptions, is structured as follows.

```bash
field-name1:type1:description1|field-name2:type2:description2|...n # for example, 'result:integer:the result of the review|reason:string:the justification of the result of the review'
```

Providing a description can be useful for both the LLM and the user in understanding the purpose of the field. It can also reduce the amount of guidance needed in the main prompt itself to ensure response content is correctly assigned.

To have the pattern be interpreted as a template for the elements of an array, rather than a singular response item, prefix the definition with `[]`,  as shown below.

```bash
[]field-name:type # for example, an array of elements, each of the form 'result:integer'...n #
```

A simple example of executing `gen` with a `GSL` defined schema is shown below.

```bash
gen --quiet --schema '[]colour:string' "list all colours of the rainbow"
```

This will return a response similar to the following.

```json
[
 {"colour": "Red"},
 {"colour": "Orange"},
 {"colour": "Yellow"},
 {"colour": "Green"},
 {"colour": "Blue"},
 {"colour": "Indigo"},
 {"colour": "Violet"}
]
```

##### Open API Schema

For more complex schemas, the definition can be provided as an [OpenAPI Schema Object](https://spec.openapis.org/oas/v3.0.3#schema-object-examples). A simple example is shown below.

```bash
gen --quiet --schema '{"type":"object","properties":{"colour":{"type":"string", "description":"the selected colour"}}}' "pick a colour of the rainbow"
```

This will return a response similar to the following.

```json
{
 "colour": "Blue"
}
```

It may be preferable to store complex `schemas` in a file rather than declaring them inline. Standard `command substitution` techniques can be used to enable this. The example below shows how the same `schema` as defined inline above can instead be read from the file `./schema.json`.

```bash
gen --schema "$(cat ./schema.json)" "pick a colour of the rainbow"
```

##### Example

The following example describes how to use `gen` to perform a basic code review of a given file and return the result in a specific, consistent `json` format. Making it suitable for use in automation.

For clarity, note the below command...

```bash
# start a new gen session in --quiet mode (suppresses output, -q can also be used). 
# - include the --files (main.go, -f can also be used) in the prompt 
# - specify a --schema in order to produce a structured response (-s can also be used)
gen --quiet \
    --files ./main.go \
    --schema 'quality:integer:1 excellent, 5 terrible|reason:string:brief justification for the quality' \
  "perform a code review on this file"
```

... will result in the following...

```json
{
 "quality": 1,
 "reason": "The code is excellent. It is well-structured with clear separation of concerns into distinct packages (cli, gemini, log, schema, session). Error handling is robust and consistent, using a custom `log.FatalfIf` helper and a panic handler. Configuration is managed cleanly by populating a `Config` struct from parsed arguments. The logic is straightforward and easy to follow for a command-line application."
}
```

Given this understanding, the script below demonstrates how this data can be used as the basis for decisions in CI, scripts or other automation, see the revised version below.

```bash
# perform the 'code review' and store the JSON response in a variable
JSON=$(gen -q -f ./main.go --schema 'quality:integer:1 excellent, 5 terrible|reason:string:brief justification for the quality grade' "perform a code review on this file") 

# parse the JSON into an array containing either the 'suggested improvements' or 'ok' and the associated exit code based on whether it was an 'ok' result or it required revising
# - eg '[ 'ok', 0 ] or '[ 'horrific stuff, unreadable...', 1 ]
RESULT=$(echo "$JSON" | jq -r 'if .quality > 2 then [.reason, 1] else ["ok", 0] end') 

# print either 'ok' or the suggested improvements
echo "$RESULT" | jq -r '.[0]'  

# exit with a return code derived from whether the minimum required quality was met
exit "$(echo "$RESULT" | jq -r '.[1]')"
```

### System Prompt Configuration

The default system prompt provides the `Gemini API` with the context that it is being prompted from a terminal process on your specific `OS` and to format responses appropriately.

Should you wish to change this, pass a `--system-prompt` value along with the prompt to `gen`, as shown below.

```bash
gen --system-prompt "you are a voice controlled toaster convinced that everyone always wants you to make them a slice of toast, regardless of the question" "what is the weather like?"
# >> Right then, coming right up – a perfect slice of toast! You know, a nice, warm, crispy piece of toast can really brighten your day, no matter what the weather is doing!
```

If, as is typically the case, the default system prompt is suitable, but you wish to add further context to it about the user or use-case, pass the `--use-case` value along with the prompt to `gen`, as shown below.

```bash
gen --use-case "you are assisting a go/linux software engineer based in staffordshire" "where might I get a job nearby?"
# >> Okay, here are some options for Go/Linux software engineer jobs in and around Staffordshire... (response truncated for brevity)
```

Any such information you provide is only used to augment requests sent to the `Gemini API`. It is not stored or transmitted in any other form or for any other purpose, under any circumstances.

#### Configuring Defaults

As with any command executed via a `posix` compliant shell, `redirection` can be used to facilitate reading arguments from files and `aliases` used to assign defaults arguments.

An example is shown below that combines both methods to configure a shell in which running `gen` applies a custom system prompt that, for the sake of illustration, is defined in a shared file. Such files could be used to share common system prompts, for instance. If such a feature is not needed, the system-prompt could simply be defined in a variable or hard coded into the alias.

```bash
# file: ~/.bashrc

# create a shared system prompt file. this would likely be fetched from a remote endpoint in reality, but it is written directly here by way of example
# alternatively, this data could simply be assigned to, and read from, a variable or included in the alias directly. whichever is more readable for the given system-prompt size
echo "you are a linux terminal based assistant running on systems owned by global-mega-corp. ensure any advice prioritises potential profit over all else. including human life." > ~/gen-system-prompt.txt
# define an alias to apply a common use-case to all local invocations and to specify the custom system prompt based on the contents of the system prompt file
alias gen='gen --system-prompt "$(cat ~/gen-system-prompt.txt)" --use-case "you are running in vs-code integrated terminal on development machine providing interactive feedback to a go/linux developer"'
```

Users of the shell can then just run `gen` directly and implicitly use the custom system prompt. As shown below.

```bash
gen "in what context are we operating?"
# >> We are operating within the integrated terminal of VS Code on your development machine. My role is to function as a Linux terminal-based assistant for you, a Go/Linux developer at Global-Mega-Corp.... (response truncated for brevity)
```

#### Configuring Multiple Modes

By making further use of `aliases`, `gen` can be pre-configured for use in different `modes`.

In the example below, two modes are configured, `architect` and `chat`

```bash
# create an architect "mode"
alias gen-arc='gen --use-case "you are a wise and beardy architect with expertise in gcp, k8s, linux and go" --pro --temperature 0.1' 

# create an general chat "mode"
alias gen-chat='gen --use-case "you are a helpful technical assistant" --temperature 0.3' 
```
Users of the shell can then just run `gen-arc` or `gen-chat` directly and access `gen` in that specific `mode`. As shown below.

```bash
gen-arc "design me a facebook, give me the result as a mermaid chart" # ask an architecture question
# >> a design for a facebook-like system would... (response truncated for brevity)

gen-chat "what features were added in go 1.24?" # ask a general question
# >> Go 1.24 introduced several new features and enhancements across the... (response truncated for brevity)
```

### Grounding

Grounding is the term for verifying Gemini's responses with an external source, that source being `Google Search` in the case of `gen`. By default, this feature is enabled, but it can be disabled with the `--no-grounding`flag, as shown below.

```bash
gen --no-grounding "how do I list all files in my current directory?"
```

## Model Configuration

By default, `gen` uses the latest `flash` model at the time of release. By passing the `--pro` flag, you can switch to using the latest thinking model (at the time of release). You can also set the model directly, along with `temperature`, `top-p` and `token limits`. An example is shown below.

```bash
gen --model 'custom-gemini-exp-model-123' --temperature 0.1 --top-p 0.1 --max-tokens=1000 "how do I list all files in my current directory?"
```

The effect of the above will be to make the responses more deterministic and favour correctness over 'imagination'.

While the effects of `top-p` and `temperature` are out of the scope of this document, briefly and simplistically; when the LLM is selecting the next token to include in its response, the value of `top-p` restricts the pool of potential next tokens that can be selected to the most probable subset. This is derived by selecting the most probable, one by one, until the cumulative probability of that selection exceeds the value of `p`. The `temperature` value is then used to weight the probabilities in that resulting subset to either level them out or emphasise their differences; making it less or more likely that the highest probability candidate will be chosen.

## Reporting on Usage

Running `gen` with the `--stats` flag will cause usage data to be written to `stderr`. This allows it to be processed separately from the main response. An example is shown below.

```bash
gen --stats "what is the weather like in london next week?"
```

This will produce output similar to the below

```bash
{
  "stats": {
    "filesStored": "0",
    "functionCall": "false",
    "model": "gemini-2.5-flash",
    "promptBytes": "45",
    "responseBytes": "1190",
    "systemPromptBytes": "761",
    "tokens": "929"
  }
}

The weather will be very hot next week
```

To redirect the `stats` component to a file, use standard redirection techniques, such as in the below example, where `stderr` is redirected to a local file.

```bash
gen --stats "what is the weather like in london next week?" 2> stats.txt
```

This will produce output similar to the below

```
The weather will be very hot next week
```

And the contents of `stats.txt` will be similar to the following.

```json
{
  "stats": {
    "filesStored": "0",
    "functionCall": "false",
    "model": "gemini-2.5-flash",
    "promptBytes": "45",
    "responseBytes": "1190",
    "systemPromptBytes": "761",
    "tokens": "929"
  }
}
```

## Debugging

To inspect the underlying Gemini API traffic that is generated by `gen`, run it with the `--verbose` (or `-v`) flag. Other arguments can be passed normally. With the `--verbose` flag specified, the `Gemini API` request and response payloads and other relevant data will be written to `stderr`. This output is in the form of JSON encoded structured logs. As the primary responses are written to `stdout` the debug component can easily be separated from the main content, for independent analysis, using standard `redirection` techniques.

