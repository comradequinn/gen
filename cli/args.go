package cli

import (
	"flag"
	"fmt"
	"os"
	"path"
	"runtime"
)

// Args defines all command line arguments
type Args struct {
	script, scriptShort                       *bool
	files, filesShort                         *string
	newSession, newSessionShort               *bool
	listSessions, listSessionsShort           *bool
	restoreSession, restoreSessionShort       *int
	deleteSession, deleteSessionShort         *int
	vertexAccessToken, vertexAccessTokenShort *string
	gcpProject, gcpProjectShort               *string
	gcsBucket, gcsBucketShort                 *string
	schemaDefinition, schemaDefinitionShort   *string
	commandExecution, commandExecutionShort   *bool
	commandApproval, commandApprovalShort     *bool
	debug, debugShort                         *bool
	CustomURL                                 *string
	CustomUploadURL                           *string
	Version                                   *bool
	DeleteAllSessions                         *bool
	DisableGrounding                          *bool
	Stats                                     *bool
	AppDir                                    *string
	CustomModel                               *string
	ProModel                                  *bool
	MaxTokens                                 *int
	Temperature                               *float64
	TopP                                      *float64
	SystemPrompt                              *string
	UseCase                                   *string
}

func ReadArgs(homeDir, app, proModel string) Args {
	args := Args{}

	args.Version = flag.Bool("version", false, "print the version")
	args.script, args.scriptShort = flagDef(flag.Bool, "script", "q", "quiet the output. supress activity indicators, such as spinners, to better support piping stdout into other utils when scripting", false)
	args.files, args.filesShort = flagDef(flag.String, "files", "f", "a comma separated list of files to attach to the prompt", "")
	args.newSession, args.newSessionShort = flagDef(flag.Bool, "new", "n", "save any existing session and start a new one", false)
	args.listSessions, args.listSessionsShort = flagDef(flag.Bool, "list", "l", "list all sessions by id", false)
	args.restoreSession, args.restoreSessionShort = flagDef(flag.Int, "restore", "r", "the session id to restore", 0)
	args.deleteSession, args.deleteSessionShort = flagDef(flag.Int, "delete", "d", "the session id to delete", 0)

	args.CustomURL = flag.String("url", "", "a custom url to use for the gemini api. by default the vertex-ai (gcp) or generative-language-api (ai-studio) canonical urls are used depending on whether "+
		"an access-token is specified or not. where no access-token is specified, the generative-language-api form is used and the GEMINI_API_KEY envar is queried for the api-key to include in its querystring. where an "+
		"access-token is specified, the vertex-ai form is used and both -gcp-project and -gcs-bucket arguments must be also specified. the following placeholders are supported in custom urls and will be populated "+
		"where specified and appropriate: {model}, {api-key}, {gcp-project}")

	args.CustomUploadURL = flag.String("upload-url", "", "a custom url to use for file uploads. by default the cloud storage (gcp) or generative-language-api (ai-studio) canonical urls are used depending on whether "+
		"an access-token is specified or not. where no access-token is specified, the generative-language-api form is used and the GEMINI_API_KEY envar is queried for the api-key to include in its querystring. where an "+
		"access-token is specified, the cloud storage form is used and both -gcp-project and -gcs-bucket arguments must be also specified. the following placeholders are supported in custom urls and will be populated "+
		"where specified and appropriate: {api-key}, {gcs-bucket}, {file-name}")

	args.vertexAccessToken, args.vertexAccessTokenShort = flagDef(flag.String, "vertex-access-token", "a", "the access token to present to the vertex-ai (gcp) gemini api endpoint. specifying a vertex-access-token will cause the "+
		"vertex-ai (gcp) canonical endpoint to be used (unless a custom url is provided)", "")

	args.gcpProject, args.gcpProjectShort = flagDef(flag.String, "gcp-project", "p", "the gcp project to include in the gemini api url. specifying a gcp-project will cause the vertex-ai (gcp) canonical endpoint to be "+
		"used (unless a custom url is provided)", "")

	args.gcsBucket, args.gcsBucketShort = flagDef(flag.String, "gcs-bucket", "b", "the cloud storage (gcp) bucket to upload files to when using the gemini api via a vertex-ai (gcp) endpoint", "")

	args.commandExecution, args.commandExecutionShort = flagDef(flag.Bool, "exec", "x", fmt.Sprintf("whether to enable command execution. when enabled prompts should relate to interacting with the local host environment "+
		"in some form. responses will typically result in %v executing commands on behalf of the gemini api", app), false)

	args.commandApproval, args.commandApprovalShort = flagDef(flag.Bool, "approve", "k", "whether to prompt for review and approval before executing commands on behalf of the gemini api", false)

	args.schemaDefinition, args.schemaDefinitionShort = flagDef(flag.String, "schema", "s", "a schema that defines the required response format. either in the form 'field1:field1-type:field1-description|field2:field2-type:field2-description|...n' or "+
		"as a json-form open-api schema. grounding with search must be disabled to use a schema", "")

	args.DeleteAllSessions = flag.Bool("delete-all", false, "delete all session data")
	args.DisableGrounding = flag.Bool("no-grounding", false, "disable grounding with search")
	args.debug, args.debugShort = flagDef(flag.Bool, "verbose", "v", "enable verbose output to support debugging", false)
	args.Stats = flag.Bool("stats", false, "print count of tokens used")
	args.AppDir = flag.String("app-dir", path.Join(homeDir, "."+app), fmt.Sprintf("location of the %v app directory", app))
	args.CustomModel = flag.String("model", "", "the specific model to use")
	args.ProModel = flag.Bool("pro", false, fmt.Sprintf("use the thinking %v model", proModel))
	args.MaxTokens = flag.Int("max-tokens", 65536, "the maximum number of tokens to allow in a response")
	args.Temperature = flag.Float64("temperature", 0.1, "the temperature setting for the model")
	args.TopP = flag.Float64("top-p", 0.1, "the top-p setting for the model")

	args.SystemPrompt = flag.String("system-prompt",
		fmt.Sprintf("You are a command line utility named '%v' running in a terminal on the OS '%v' with a locale set to '%v'. Factor that into the format and content of your responses and always ensure they are concise and "+
			"easily rendered in such a terminal. Do not use complex markdown syntax in your responses as this is not rendered well in terminal output. Do use clear, plain text formatting that can be easily read "+
			"by a human; such as using dashes for list delimiters. Always ensure that, to the extent that you are reasonably able, that your answers are factually correct and you take caution regarding hallucinations. "+
			"Only answer the specific question given and do not proactively include additional information that is not directly relevant to that question. ", app, runtime.GOOS, os.Getenv("LANG")),
		"the system prompt to use")

	args.UseCase = flag.String("use-case", "", "free text information to include in the system prompt about the user or use-case, such as a role or location. "+
		"for example 'you are running in a ci pipeline used to verify code quality' or 'you are assisting a go/linux software engineer based in staffordshire'")

	flag.Parse()

	return args
}

func (args Args) Script() bool {
	return readFlag("script/q", args.script, args.scriptShort)
}

func (args Args) Files() string {
	return readFlag("files/f", args.files, args.filesShort)
}

func (args Args) NewSession() bool {
	return readFlag("new/n", args.newSession, args.newSessionShort)
}

func (args Args) ListSessions() bool {
	return readFlag("list/l", args.listSessions, args.listSessionsShort)
}

func (args Args) RestoreSession() int {
	return readFlag("restore/r", args.restoreSession, args.restoreSessionShort)
}

func (args Args) DeleteSession() int {
	return readFlag("delete/d", args.deleteSession, args.deleteSessionShort)
}

func (args Args) VertexAccessToken() string {
	return readFlag("vertex-access-token/a", args.vertexAccessToken, args.vertexAccessTokenShort)
}

func (args Args) GCPProject() string {
	return readFlag("gcp-project/p", args.gcpProject, args.gcpProjectShort)
}

func (args Args) GCSBucket() string {
	return readFlag("gcs-bucket/b", args.gcsBucket, args.gcsBucketShort)
}

func (args Args) CommandExecution() bool {
	return readFlag("exec/x", args.commandExecution, args.commandExecutionShort)
}

func (args Args) CommandApproval() bool {
	return readFlag("approve/k", args.commandApproval, args.commandApprovalShort)
}

func (args Args) SchemaDefinition() string {
	return readFlag("schema/s", args.schemaDefinition, args.schemaDefinitionShort)
}

func (args Args) Debug() bool {
	return readFlag("verbose/v", args.debug, args.debugShort)
}

// flagDef is a helper function to define a flag and its shortform
func flagDef[T any](flagFunc func(string, T, string) *T, name, shortform, desc string, val T) (*T, *T) {
	f := flagFunc(name, val, desc)
	fs := flagFunc(shortform, val, "shortform of -"+name)

	return f, fs
}

// readFlag is a helper function to read a flag or its shortform
func readFlag[T comparable](name string, longform, shortform *T) T {
	var defaultT T

	if *longform != defaultT && *shortform != defaultT {
		panic("both longform and shortform variants were provided for the flag '" + name + "'. provide one or the other")
	}

	if *longform != defaultT {
		return *longform
	}

	return *shortform
}
