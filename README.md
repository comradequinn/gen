# gen

Named `gen` (from `generate`), `gen` is an agentic, command-line `llm` interface built on Google's `Gemini` models. 

Using `gen` greatly simplifies integrating LLM capabilities into CI pipelines, scripts or other automation. 

For terminal users, `gen` acts as a fully-featured conversational, agentic assistant.

## Features

Using `gen` provides the following features:

* Agentic, conversational, command-line chatbot
  * Non-blocking, yet conversational, prompting allowing natural, fluid usage within the terminal environment
    * The avoidance of a dedicated `repl` to define a session leaves the terminal free to execute other commands between prompts while still maintaining the conversational context
  * Agentic features with `exec` mode
    * Ask `gen` to `do` a task for you rather than explain how
      * Query file contents, git repos and remote APIs
      * Analyse data and write the results to new or existing files
      * Install programs, download files and scrape websites
      * Perform complex multi-stage tasks with a single prompt
  * Session management enables easy stashing of, or switching to, the currently active, or a previously stashed session
    * This makes it simple to quickly task switch without permanently losing the current conversational context
* Fully scriptable and ideal for use in automation and CI pipelines
  * All configuration and session history flag or file based
  * API Keys are provided via environment variables or flags
  * Support for structured responses using custom `schemas`
    * Basic schemas can be defined using a simple schema definition language
    * Complex schemas can be defined using OpenAPI Schema objects expressed as JSON (either inline or in dedicated files)
  * Interactive-mode activity indicators can be disabled to aid effective redirection and piping
* Support for attaching one or many files to prompts
  * Interrogate individual code, markdown and text files or entire workspaces
  * Describe image files and PDFs
* System prompt configuration
  * Specify general and user/use-case based system prompt content
* Model configuration
  * Specify custom model configurations to fine-tune output

## Installation

To install `gen`, download the appropriate tarball for your `os` from the [releases](https://github.com/comradequinn/gen/releases/) page. Extract the binary and place it somewhere accessible to your `$PATH` variable. 

Optionally, you can use the below script to do that for you

```bash
export VERSION="v1.3.3"; export OS="linux-amd64"; wget "https://github.com/comradequinn/gen/releases/download/${VERSION}/gen-${VERSION}-${OS}.tar.gz" && tar -xf "gen-${VERSION}-${OS}.tar.gz" && rm -f "gen-${VERSION}-${OS}.tar.gz" && chmod +x gen && sudo mv gen /usr/local/bin/
```

### Authentication

You can configure `gen` to access the `Gemini API` either via the publicly available `Generative Language API` endpoints (as used by `Google AI Studio`) or via a `Vertex AI` endpoint managed within a `Google Cloud Platform (GCP)` project.

#### Generative Language API (Google AI Studio)

To use `gen` via the `Generative Language API`, set and export your `Gemini API Key` as the conventional environment variable for that value: `GEMINI_API_KEY`.

If you do not already have a `Gemini API Key`, they are available free from [Google AI Studio](https://aistudio.google.com), [here](https://aistudio.google.com/apikey). 

For convenience, you may wish to add the envar definition to your `~/.bashrc` file. An example of doing this is shown below. 

```bash
# file: ~/.bashrc

export GEMINI_API_KEY="myPriVatEApI_keY_1234567890"
```

Remember that you will need to open a new terminal or `source` the `~/.bashrc` file for the above to take effect.

Once this is done, `gen` will default to using the `Generative Language API` and your `GEMINI_API_KEY` unless you explicitly specifiy `Vertex AI (Google Cloud Platform)` credentials to use instead; in which case they will take precedence.

#### Vertex AI (Google Cloud Platform)

To use `gen` with a `Vertex AI` `Gemini API` endpoint, firstly configure `ADC (application default credentials)` on your workstation, if you have not already done so, by running the below.

```bash
gcloud auth application-default login --disable-quota-project
```

You can then render `access tokens` using `gcloud auth application-default print-access-token`. These can be passed to `gen` using a `--vertex-access-token` (or `-a`) argument. A `GCP Project` and a `GCS Bucket` must also be specified, using `--gcp-project` (or `-p`) and `--gcs-bucket` (or `-b`), respectively. An example is shown below.

```bash
gen --new --vertex-access-token "$(gcloud auth application-default print-access-token)" --gcp-project "my-project" --gcs-bucket "my-bucket" "what is the weather like in London tomorrow?"
```

When you specify `Vertex AI` credentials, they take precedence over any `GEMINI_API_KEY` you may have set to authenticate with the `Generative Language API`.

##### Configuring Defaults

As with any command executed via a `posix` compliant shell, `aliases` can be used to assign defaults arguments. An example is shown below that configures a default `access-token` and `gcp-project`.

```bash
# file: ~/.bashrc

alias gen='gen --vertex-access-token "$(gcloud auth application-default print-access-token)" --gcp-project "my-project" --gcs-bucket "my-bucket"'
```

Users of the shell can then simply run `gen` directly and implicitly use the those `gcp credentials`. As shown below.

```bash
gen -n "what is the weather like in London tomorrow?"
```

### Removal

To remove `gen`, delete the binary from `/usr/bin` (or the location it was originally installed to). You may also wish to delete its application directory. This stores user preferences and session history and is located at `~/.gen`.

## Quick Start

These examples show typical sequences for the two forms of `gen` usage: interactive and scripted. For those familiar with the command line, scripting and LLMs, they are likely enough you to become productive with `gen`.

### Interactive Mode

The following example shows the fundamentals of interactive, conversational usage of `gen` within a user's terminal

```bash
gen "what is the latest version of go?" # ask a question
# >> The latest stable version of Go is.... (response ommitted for brevity)

gen "what were the major amendments in that release?" # ask a follow up question relying on the previous question for context
# >> the release introduced several significant amendments across its toolchain, runtime and... (response ommitted for brevity)

gen -n "what is the weather like in london tomorrow?" # stash the existing conversational context and start a new session (-n is the shortform of --new)
# >> In London tomorrow it will be grey and wet... (response ommitted for brevity)

gen -l # show current and active sessions, the asterix indicates the active session (-l is the shortform of --list)
   #1 (April 26 2025): what is the latest version of go?
*  #2 (April 26 2025): what is the weather like in london tomorrow?

gen -r 1 # switch the active session back to the earlier topic (-r is the shortform of --restore)

gen "was this a major or minor release?" # ask a follow up question relying on context from the restored session
# >> It was a minor release. Go follows a versioning scheme that.... (response ommitted for brevity)

gen -n -f ./main.go "summarise this code for me" # attach a file to the prompt and ask a question related to its contents (-f is the  shortform of --files)
# >> This is a Go program, named 'gen'. It functions as a command-line interface for interacting.... (response ommitted for brevity)

gen -n -x "list all .go files in my current directory" # directly execute and interpret terminal commands to perform tasks (-x is the shortform of --exec)
# >> executing... [ls -l *.go]
# >> main.go

gen -x "copy the files to a directory named 'backup'" # execute a follow up task using the previous task for context
# >> executing... [mkdir -p backup; cp *.go backup/]
# >> OK

gen -n -x "write the contents returned from github's home page to a file named github.html" # access and parse external resources 
# >> executing... [curl https://github.com > github.html]
# >> OK

gen -n -x "get the pids, names and cpu and mem usage of the top 5 processes running by cpu. format this in markdown as a table" # query the host system for performance metrics
# >> executing... [ps -eo pid,cmd,%cpu,%mem --sort=-%cpu | head -n 6]
# >> 
# >> | PID | CMD | %CPU | %MEM |
# >> |---|---|---|---|
# >> | 1331 | /home/me/process/job1 | 6.2 | 5.6 |
# >> | 6911 | /home/me/process/code | 1.4 | 0.6 |
# >> | 519 | /home/me/process/job2 | 1.1 | 0.7 |
# >> | 1420 | /home/me/process/task1 | 0.6 | 2.7 |
# >> | 2058 | /home/me/process/job3 | 0.5 | 0.4 |

gen -n -x -pro "perform a code review on the main.go file" # use the 'pro' model to handle more complex tasks
# >> This is a solid, well-written `main.go` file for a command-line application. It demonstrates good Go programming... (response ommitted for brevity)
```

### Scripting

#### Structured Responses

The following example describes how to use `gen` to perform a basic code review of a given file and return the result in a specific, consistent `json` format. Making it suitable for use in automation.

For clarity, note that the below command...

```bash
# start a new gen session in --script mode (supresses output, -q can also be used). include the --files (main.go, -f can also be used) in the prompt and specify a --schema for the response (-s can also be used)
gen --new --script --pro --files ./main.go --schema 'quality:integer:1 excellent, 5 terrible|reason:string:brief justification for the quality grade' "perform a code review on this file"
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
JSON=$(gen -n -q -f ./main.go --schema 'quality:integer:1 excellent, 5 terrible|reason:string:brief justification for the quality grade' "perform a code review on this file")  

# parse the JSON into an array containing either the 'suggested improvements' or 'ok' and the associated exit code based on whether it was an 'ok' result or it required revising
# - eg '[ 'ok', 0 ] or '[ 'horrific stuff, unreadable...', 1 ]
RESULT=$(echo "$JSON" | jq -r 'if .quality > 2 then [.reason, 1] else ["ok", 0] end')  

# print either 'ok' or the suggested improvements
echo "$RESULT" | jq -r '.[0]'   

# exit with a return code derived from whether the minimum required quality was met
exit "$(echo "$RESULT" | jq -r '.[1]')" 
```

#### Agentic Actions

The following example describes how to use `gen` to perform a agentic tasks in a script, in this case another variant on the automated code review of a given file.

```bash
gen --new --exec --script "perform a code review on the main.go file. 
write a single integer quality result value to a new file named 'result.txt', 
with 1 being 'excellent' and 5, 'terrible'. 
if the result is not 1, write a justifification as to what needs improving
to a new file named 'feedback.txt'. finally, if the result was not 1, issue a 
command to terminate the process with a return code of of 2"

case $? in
    0)
        # take whatever action is required for a completed code review with a positive outcome...
        echo "code review completed. code is a masterpiece. score $(cat result.txt)" 
        ;;
    1)
        # take whatever action is required for code review that could not be completed due to an error executing the commands...
        echo "an error prevented the code review being undertaken" 
        ;;
    2)
        # take whatever action is required for a completed code review with a negative outcome...
        echo "code review completed. code is awful; an afront to the intellectual diginity of man and beast. detailed feedback: $(cat feedback.txt)" 
        ;;
    *)
        # default error case for any other exit codes
        echo "an unknown error occurred. exit code: $?"
        ;;
esac
```

When executed, the above will cause `gen` to silently perform all the tasks and the subsequent part of the script will then print an appropriate response based on `gen's` return code. In the prompt used, the `exit code` to indicate code review failure was set as `2`; this is optional and purely to allow the two causes of failure to be distinguished in the script.

## Usage 

### Prompting

To chat with `gen`, execute it with a prompt

```bash
gen "how do I list all files in my current directory?"
```
The result of the prompt will be displayed, as shown below.

```
To list all files in your current directory, you can use the following command in your terminal:
ls -a
This command will display all files, including hidden files (files starting with a dot).
```

To ask a follow up question, run `gen` again with the required prompt.

```bash
gen "I need timestamps in the output"
```

This will return something similar to the below, note how `gen` understood the context of the question in relation to the previous prompt. 

```
To include timestamps in the output of the `ls` command, you can use the `-l` option along with the `--full-time` or `--time-style` options
```

This conversational context will continue indefinitely until you start a new session. Starting a new session `stashes` the existing conversational context and begins a new one. It is performed by passing the `--new` (or `-n`) flag in your next prompt. As shown below

```bash
gen -n "what was my last question?"
```

This will return something similar to the below, indicating the loss of the previous context.

```
I have no memory of past conversations. Therefore, I don't know what your last question was.
```

### Agentic Actions

To run `gen` in `exec` mode, pass the `--exec` (or `-x`) flag. 

When running in `exec` mode, `gen` behaves `agentically`. It is able to execute commands as your user in order to perform tasks on your behalf. These tasks can effectively be anything that you could undertake yourself and can also contain multiple steps. The exception to this is long running programs. Asking `gen` to capture all tcp traffic to a host will work, but if you do not also specify some form of exit condition, the command will run indefinitely and the result will never be collected and forwarded to `gemini` in order for it to respond; `gen` will just appear to hang.

The `--exec` flag is scoped to each invidual prompt, so agentic capabilities can be variably enabled or disabled on individual prompts within with the same conversation. 

For security purposes, it is impossible for `gen` to execute commands without the `--exec` (or `-x`) flag, and conversely, take extra caution with your prompts when `exec` mode is enabled. If you would prefer to approve each command that `gen` requests before it is executed, pass the `--approve` (or `-k`) flag along with `--exec`.

An example is shown below of using agentic mode in a conversation.

```bash
# ask for support with a task, as --exec is not specified, 
gen -n "how would I list all files in my home directory, including hidden ones, that were modified in the last day?"
# >> You can list all files in your home directory, including hidden ones, that were modified in the last day (24 hours) using the `find` command:
# >>    find ~ -maxdepth 1 -type f -mtime -1

# switch to exec mode by passing --exec flag for the next prompt and use the previous context to infer the action gen is take
gen -x "ok, execute that for me and print the results"
# >> executing... [find ~ -mtime -1]
# >> 
# >> 40344544     56 -rw-------   1 me     me        54416 Jun 12 00:08 /home/me/.bash_history
# >> 40343278      4 -rw-rw-r--   1 me     me          281 Jun 12 20:15 /home/me/.gitconfig
# >> 40344559     12 -rw-------   1 me     me        10435 Jun 12 00:07 /home/me/.viminfo
```
As well as relatively specific tasks, such as those above, you can also give general instructions to `gen`, and have it figure the steps it needs to take and execute them. As illustrated below.

```bash
gen -n -x "what kind of license is associated with this repo? can I use it in commercial software?"
# >> I need to inspect the repository's files to find a license. Executing `ls -a` to list all files, including hidden ones, in the current directory...
# >> This repository is licensed under the GNU General Public License v3.0.
# >> Key aspects of this license regarding commercial use are:
# >> - [truncated for brevity]
# >> In summary: Yes, you can use it in commercial software, but [truncated for brevity]
```

### Session Management

A session is a thread of prompts and responses with the same context, effectively a conversation. A new session starts whenever `--new` (or `-n`) is passed along with the prompt to `gen`. At this point, the previously active session is `stashed` and the passed prompt becomes the start of a new session.

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

Running `gen -l` again will now show the below; note how the asterix is now positioned at record `1`

```bash
gen -l
* #1 (April 15 2025): 'how do i list all files in my current directory?'
  #2 (April 15 2025): 'what was my last question?'
```

Asking the prompt from earlier, of `gen "what was my last question?"`, will now return the below, as that context has been restored.

```
Your last question was: "I need timestamps in the output".
```

To delete a single session, run `gen --delete #id` (or `-d #id`) where `#id` is the `#ID` in the `gen --list` output. To delete all sessions, run `gen --delete-all`

### System Prompt Configuration

The default system prompt provides the `Gemini API` with the context that it is being prompted from a terminal process on your specific `OS` and to format responses appropriately. 

Should you wish to change this, pass a `--system-prompt` value along with the prompt to `gen`, as shown below.

```bash
gen -n --system-prompt "you are a voice controlled toaster convinced that everyone always wants you to make them a slice of toast, regardless of the question" "what is the weather like?"
# >> Right then, coming right up – a perfect slice of toast! You know, a nice, warm, crispy piece of toast can really brighten your day, no matter what the weather is doing!
```

If, as is typically the case, the default system prompt is suitable, but you wish to add further context to it about the user or use-case, pass the `--use-case` value along with the prompt to `gen`, as shown below.

```bash
gen -n --use-case "you are assisting a go/linux software engineer based in staffordshire" "where might I get a job nearby?"
# >> Okay, here are some options for Go/Linux software engineer jobs in and around Staffordshire... [truncated for brevity]
```

Any such information you provide is only used to augment requests sent to the `Gemini API`. It is not stored or transmitted in any other form or for any other purpose, under any circumstances. 

#### Configuring Defaults

As with any command executed via a `posix` compliant shell, `redirection` can be used to faciliate reading arguments from files and `aliases` used to assign defaults arguments. 

An example is shown below that combines both methods to configure a shell with a custom system prompt.

```bash
# file: ~/.bashrc

# create a shared system prompt file. this would likely be fetched from a remote endpoint in reality, but it is just written directly here by way of example
echo "you are a linux terminal based assistant running on systems owned by global-mega-corp. ensure any advice prioritises potential profit over all else. including human life." > ~/gen-system-prompt.txt
# define an alias to apply a common use-case to all local invocations and to specify the custom system prompt based on the contents of the system prompt file
alias gen='gen --system-prompt "$(cat ~/gen-system-prompt.txt)" --use-case "you are running in vs-code integrated terminal on development machine providing interactive feedback to a go/linux developer"'
```

Users of the shell can then just run `gen` directly and implicitly use the custom system prompt. As shown below.

```bash
gen -n "in what context are we operating?"
# >> We are operating within the integrated terminal of VS Code on your development machine. My role is to function as a Linux terminal-based assistant for you, a Go/Linux developer at Global-Mega-Corp.... [truncated for brevity]
```

##### Configuring Multiple Modes

By making further use of `aliases`, `gen` can be pre-configured for use in different `modes`. 

In the example below, two modes are configured, `architect` and `chat`

```bash
# create an architecture "mode"
alias gen-arc='gen --use-case "you are a wise and beardy architect with expertise in gcp, k8s, linux and go" --pro --temperature 0.1'  

# create an general chat "mode"
alias gen-chat='gen --use-case "you are a helpful technical assistant" --temperature 0.3'  
```
Users of the shell can then just run `gen-arc` or `gen-chat` directly and access `gen` in that specific `mode`. As shown below.

```bash
gen-arc "design me a facebook, give me the result as a mermaid chart" # ask an architecture question
# >> a design for a facebook-like system would... (response ommitted for brevity)

gen-chat -n "what features were added in go 1.24?" # ask a general question
# >> Go 1.24 introduced several new features and enhancements across the... (response ommitted for brevity)
```

### Attaching Files

To attach files to your prompt, use the `--files` (or `-f`) parameter passing the path to the file to include. To include multiple files, separate them with a comma, Some examples are shown below.

```bash
# attach a single file
gen -n --files "holday-fun.png" "what's in this image?"
```

```bash
# attach multiple files, spaces are optional but can aid readability when listing many files explicitly. here we use the shorform of --files; -f
gen -n -f "some-code.go, somedir/some-more-code.go, yet-more-code.go" "summarise these files"
```

When attaching a large number of files or the contents of multiple directories, any `posix` compliant shell will support `command substitution` which can be used to simplify creating the `files` argument. An example is shown below of including all `*.go` files in the current workspace (that being the working directory and below).

```bash
# find all files in the current workspace and concatenate them into a single string
WORKSPACE="$(find . -name "*.go" | paste -s -d ",")"
# attach the files to the prompt
gen -n -f "$WORKSPACE" "create a table of file names and a very brief content summary for these files"
```

When run on the `gen` repo, the above will produce something similar to the below.

```text
| File | Summary |
| --- | --- |
| `main.go` | The main entry point for the `gen` command-line application. It handles argument parsing, command dispatching, and initialization. |
| `cli/args.go` | Defines and parses all command-line flags and arguments for the application using Go's `flag` package. |
| `cli/execute.go` | Contains the logic to execute shell commands requested by the Gemini model, including handling user approval for security. |
| `cli/generate.go` | Orchestrates the core 'generate' functionality, managing the prompt-response loop, session history, and function/tool execution. |
| `cli/io.go` | Provides utility functions for command-line I/O, such as writing formatted output and displaying a loading spinner. |
| `cli/list.go` | Implements the functionality to display a formatted list of saved and active user sessions. |
| `gemini/gemini.go` | The primary file for the `gemini` package, containing the `Generate` function that constructs and sends requests to the Gemini API. |
| `gemini/config.go` | Defines the `Config` struct for holding all configuration parameters for the Gemini client and sets default values. |
| `gemini/config_test.go` | Contains unit tests for the configuration logic, ensuring that default values and constraints are applied correctly. |
| `gemini/history.go` | Manages the construction of conversation history to be sent with each new prompt, converting past transactions into the required format. |
| `gemini/http.go` | Handles the low-level HTTP request and response interaction with the Gemini API endpoint. |
| `gemini/tools.go` | Defines the tools available to the model, such as command execution and file requests, and their corresponding JSON schema for the API. |
| `gemini/types.go` | Defines the core data structures for the application, such as `Prompt`, `Transaction`, `Input`, and `Output`. |
| `gemini_test.go` | Contains unit tests for the main `gemini.Generate` function, using mock servers to simulate API interactions. |
| `gemini/internal/resource/resource.go` | Provides a generic interface and utilities for file handling, including concurrent batch uploads and MIME type detection. |
| `gemini/internal/resource/gcs/gcs.go` | Implements file uploads specifically for Google Cloud Storage (GCS), for use with Vertex AI. |
| `gemini/internal/resource/gla/gla.go` | Implements the resumable upload protocol for Google's Generative Language API (GLA) file service. |
| `gemini/internal/schema/schema.go` | Contains logic to build a JSON OpenAPI schema from a simplified string definition provided by the user. |
| `gemini/internal/schema/schema_test.go`| Unit tests for the schema-building logic, verifying correct JSON output for various valid and invalid definitions. |
| `gemini/internal/schema/types.go` | Defines the Go structs that map to the JSON request and response schemas of the Gemini API. |
| `log/log.go` | A simple logging utility that wraps the standard `slog` package for application-wide logging. |
| `session/session.go` | Manages session state, including reading, writing, stashing, restoring, and deleting conversation history from the file system. |
| `session/session_test.go` | Contains unit tests for the session management functionality, ensuring that session state can be manipulated correctly. |
```

### Grounding

Grounding is the term for verifying Gemini's responses with an external source, that source being `Google Search` in the case of `gen`. By default this feature is enabled, but it can be disabled with the `--no-grounding`flag, as shown below.

```bash
gen -n --no-grounding "how do I list all files in my current directory?"
```

### Scripting

When using the output of `gen` in a script, it is advisable to supress activity indicators and other interactive output using the `--script` flag (or `-q`). This ensures a consistent output stream containing only response data.

The simple example below uses redirection to write the response to a file.

```bash
gen -n --script "pick a colour of the rainbow" > colour.txt
```
This will result in a file similar to the below

```bash
# file: colour.txt
Blue
```

#### Agentic Actions in Scripts

Using `exec` mode in scripts is largely the same as using it conversationally. The key difference is that there is no user interaction. So `gen` cannot seek instruction on how to proceed after an error. Similarly when commands executed by `gen` fail, a script would typically expect `gen` to terminate and pass that exit code back to it; which is indeed how it behaves.

You can also instruct `gen` to exit with a specific return codes in your prompt too, an example is shown below.

```bash
gen -n --script -exec "ascertain the day of the week and then terminate the process with exit code 1 if it monday, 2 if is tuesday, or 3 otherwise"
##> OK

echo "$?"
# 3 (assuming it is not monday or tuesday)
```

### Structured Responses

By default, `gen` will request responses structured as free-form text, which is a sensible format for conversational use. However, in many scenarios, particularly CI and scripting use-cases, it is preferable to have the output in a structured form. To this end, `gen` allows you to specify a schema, using `--schema` (or `-s`), that will be used to structure the response.

There are two methods of specifying a schema, either by using `GSL` (`gen`'s `s`chema `l`anguage) or by providing a JSON based `OpenAPI schema object`. 

In either case, note that `grounding` will be implicitly disabled when using a `schema`, this is a current stipulation of the `Gemini API`, not `gen` itself.

#### GSL (Gen's Schema Language)

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
gen -n --script --schema '[]colour:string' "list all colours of the rainbow"
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

#### Open API Schema

For more complex schemas, the definition can be provided as an [OpenAPI Schema Object](https://spec.openapis.org/oas/v3.0.3#schema-object-examples). A simple example is shown below.

```bash
gen -n --script --schema '{"type":"object","properties":{"colour":{"type":"string", "description":"the selected colour"}}}' "pick a colour of the rainbow"
```

This will return a response similar to the following.

```json
{
  "colour": "Blue"
}
```

It may be preferable to store complex `schemas` in a file rather than declaring them inline. Standard `command substitution` techniques can be used to enable this. The example below shows how the same `schema` as defined inline above can instead be read from the file `./schema.json`.

```bash
gen -n --schema "$(cat ./schema.json)" "pick a colour of the rainbow"
```

## Model Configuration 

By default, `gen` uses the latest `flash` model at the time of release. By passing the `--pro` flag, you can switch to using the latest thinking model (at the time of release). You can also set the model directly, along with `temperature`, `top-p` and `token limits`. An example is shown below.

```bash
gen --model 'custom-gemini-exp-model-123' --temperature 0.1 --top-p 0.1 --max-tokens=1000 "how do I list all files in my current directory?"
```

The effect of the above will be to make the responses more deterministic and favour correctness over 'imagination'. 

While the effects of `top-p` and `temperature` are out of the scope of this document, briefly and simplistically; when the LLM is selecting the next token to include in its response, the value of `top-p` restricts the pool of potential next tokens that can be selected to the most probable subset. This is derived by selecting the most probable, one by one, until the cumulative probability of that selection exceeds the value of `p`. The `temperature` value is then used to weight the probabilities in that resulting subset to either level them out or emphasise their differences; making it less or more likely that the highest probability candidate will be chosen.

## Reporting on Usage

Running `gen` with the `--stats` flag will cause usage data to be written to `stderr`. This allows it be processed separately from the main response. An example is shown below.

```bash
gen -n --stats "what is the weather like in london next week?"
```

This will produce output similar to the below

```bash
The weather will be very hot next week

{
  "stats": {
    "filesStored": "0",
    "functionCall": "false",
    "model": "gemini-2.5-flash",
    "promptBytes": "45",
    "responseBytes": "1077",
    "systemPromptBytes": "761",
    "tokens": "757"
  }
}
```

To redirect the `stats` component to a file, use standard redirection techniques, such as in the below example, where `stderr` is redirected to a local file.

```bash
gen -n --stats "what is the weather like in london next week?" 2> stats.txt
```

This will produce output similar to the below

```
The weather will be very hot next week
```

And the contents of `stats.txt` will be similar to the following.

```json
{
  "stats": {
    "files": "0",
    "functionCall": "false",
    "model": "gemini-2.5-flash",
    "promptBytes": "45",
    "responseBytes": "1077",
    "systemPromptBytes": "761",
    "tokens": "757"
  }
}
```

## Debugging

To inspect the underlying Gemini API traffic that is generated by `gen`, run it with the `--verbose` (or `-v`) flag. Other arguments can be passed normal. With the `--verbose` flag specified, the `Gemini API` request and response payloads and other relevant data will be written to `stderr`. This output is in the form of JSON encoded structured logs. As the primary responses are written to `stdout` the debug component can easily be separated from the main content, for independent analysis, using standard `redirection` techniques.