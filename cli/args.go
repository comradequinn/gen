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
	Version                                   *bool
	Script, ScriptShort                       *bool
	Files, FilesShort                         *string
	NewSession, NewSessionShort               *bool
	ListSessions, ListSessionsShort           *bool
	RestoreSession, RestoreSessionShort       *int
	DeleteSession, DeleteSessionShort         *int
	CustomURL, CustomUploadURL                *string
	VertexAccessToken, VertexAccessTokenShort *string
	GCPProject, GCPProjectShort               *string
	GCSBucket, GCSBucketShort                 *string
	SchemaDefinition                          *string
	SchemaDefinitionShort                     *string
	CommandExecution                          *bool
	CommandExecutionShort                     *bool
	CommandApproval                           *bool
	CommandApprovalShort                      *bool
	DeleteAllSessions                         *bool
	DisableGrounding                          *bool
	Debug, DebugShort                         *bool
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
	args.Script, args.ScriptShort = flagDef(flag.Bool, "script", "q", "quiet the output. supress activity indicators, such as spinners, to better support piping stdout into other utils when scripting", false)
	args.Files, args.FilesShort = flagDef(flag.String, "files", "f", "a comma separated list of files to attach to the prompt", "")
	args.NewSession, args.NewSessionShort = flagDef(flag.Bool, "new", "n", "save any existing session and start a new one", false)
	args.ListSessions, args.ListSessionsShort = flagDef(flag.Bool, "list", "l", "list all sessions by id", false)
	args.RestoreSession, args.RestoreSessionShort = flagDef(flag.Int, "restore", "r", "the session id to restore", 0)
	args.DeleteSession, args.DeleteSessionShort = flagDef(flag.Int, "delete", "d", "the session id to delete", 0)

	args.CustomURL = flag.String("url", "", "a custom url to use for the gemini api. by default the vertex-ai (gcp) or generative-language-api (ai-studio) canonical urls are used depending on whether "+
		"an access-token is specified or not. where no access-token is specified, the generative-language-api form is used and the GEMINI_API_KEY envar is queried for the api-key to include in its querystring. where an "+
		"access-token is specified, the vertex-ai form is used and both -gcp-project and -gcs-bucket arguments must be also specified. the following placeholders are supported in custom urls and will be populated "+
		"where specified and appropriate: {model}, {api-key}, {gcp-project}")

	args.CustomUploadURL = flag.String("upload-url", "", "a custom url to use for file uploads. by default the cloud storage (gcp) or generative-language-api (ai-studio) canonical urls are used depending on whether "+
		"an access-token is specified or not. where no access-token is specified, the generative-language-api form is used and the GEMINI_API_KEY envar is queried for the api-key to include in its querystring. where an "+
		"access-token is specified, the cloud storage form is used and both -gcp-project and -gcs-bucket arguments must be also specified. the following placeholders are supported in custom urls and will be populated "+
		"where specified and appropriate: {api-key}, {gcs-bucket}, {file-name}")

	args.VertexAccessToken, args.VertexAccessTokenShort = flagDef(flag.String, "vertex-access-token", "a", "the access token to present to the vertex-ai (gcp) gemini api endpoint. specifying a vertex-access-token will cause the "+
		"vertex-ai (gcp) canonical endpoint to be used (unless a custom url is provided)", "")

	args.GCPProject, args.GCPProjectShort = flagDef(flag.String, "gcp-project", "p", "the gcp project to include in the gemini api url. specifying a gcp-project will cause the vertex-ai (gcp) canonical endpoint to be "+
		"used (unless a custom url is provided)", "")

	args.GCSBucket, args.GCSBucketShort = flagDef(flag.String, "gcs-bucket", "b", "the cloud storage (gcp) bucket to upload files to when using the gemini api via a vertex-ai (gcp) endpoint", "")

	args.CommandExecution, args.CommandExecutionShort = flagDef(flag.Bool, "exec", "x", fmt.Sprintf("whether to enable command execution. when enabled prompts should relate to interacting with the local host environment "+
		"in some form. responses will typically result in %v executing commands on behalf of the gemini api", app), false)

	args.CommandApproval, args.CommandApprovalShort = flagDef(flag.Bool, "approve", "k", "whether to prompt for review and approval before executing commands on behalf of the gemini api", false)

	args.SchemaDefinition, args.SchemaDefinitionShort = flagDef(flag.String, "schema", "s", "a schema that defines the required response format. either in the form 'field1:field1-type:field1-description|field2:field2-type:field2-description|...n' or "+
		"as a json-form open-api schema. grounding with search must be disabled to use a schema", "")

	args.DeleteAllSessions = flag.Bool("delete-all", false, "delete all session data")
	args.DisableGrounding = flag.Bool("no-grounding", false, "disable grounding with search")
	args.Debug, args.DebugShort = flagDef(flag.Bool, "verbose", "v", "enable verbose output to support debugging", false)
	args.Stats = flag.Bool("stats", false, "print count of tokens used")
	args.AppDir = flag.String("app-dir", path.Join(homeDir, "."+app), fmt.Sprintf("location of the %v app directory", app))
	args.CustomModel = flag.String("model", "", "the specific model to use")
	args.ProModel = flag.Bool("pro", false, fmt.Sprintf("use the thinking %v model", proModel))
	args.MaxTokens = flag.Int("max-tokens", 10000, "the maximum number of tokens to allow in a response")
	args.Temperature = flag.Float64("temperature", 0, "the temperature setting for the model")
	args.TopP = flag.Float64("top-p", 0.2, "the top-p setting for the model")

	args.SystemPrompt = flag.String("system-prompt",
		fmt.Sprintf("You are a command line utility named '%v' running in a terminal on the OS '%v' with a locale set to '%v'. Factor that into the format and content of your responses and always ensure they are concise and "+
			"easily rendered in such a terminal. You do not use complex markdown syntax in your responses as this is not rendered well in terminal output. You do use clear, plain text formatting that can be easily read "+
			"by a human; such as using dashes for list delimiters. You always ensure that, to the extent that you are reasonably able, that your answers are factually correct and you take caution regarding hallucinations. "+
			"You only answer the specific question given and do not proactively include additional information that is not directly relevant to that question. ", app, runtime.GOOS, os.Getenv("LANG")),
		"the system prompt to use")

	args.UseCase = flag.String("use-case", "", "free text information to include in the system prompt about the user or use-case, such as a role or location. "+
		"for example 'you are running in a ci pipeline used to verify code quality' or 'you are assisting a go/linux software engineer based in staffordshire'")

	flag.Parse()

	return args
}

// flagDef is a helper function to define a flag and its shortform
func flagDef[T any](flagFunc func(string, T, string) *T, name, shortform, desc string, val T) (*T, *T) {
	f := flagFunc(name, val, desc)
	fs := flagFunc(shortform, val, "shortform of -"+name)

	return f, fs
}
