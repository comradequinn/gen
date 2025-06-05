package cli

import (
	"flag"
	"fmt"
	"path"
	"runtime"
)

// Args defines all command line arguments
type Args struct {
	Version, VersionShort                     *bool
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
	DeleteAllSessions                         *bool
	DisableGrounding                          *bool
	Debug                                     *bool
	Stats                                     *bool
	AppDir                                    *string
	CustomModel                               *string
	FlashModel                                *bool
	MaxTokens                                 *int
	Temperature                               *float64
	TopP                                      *float64
	SystemPrompt                              *string
	UseCase                                   *string
}

func ReadArgs(homeDir, app, flashModel string) Args {
	flags := Args{}

	flags.Version, flags.VersionShort = flagDef(flag.Bool, "version", "v", "print the version", false)
	flags.Script, flags.ScriptShort = flagDef(flag.Bool, "script", "s", "supress activity indicators, such as spinners, to better support piping stdout into other utils when scripting", false)
	flags.Files, flags.FilesShort = flagDef(flag.String, "files", "f", "a comma separated list of files to attach to the prompt", "")
	flags.NewSession, flags.NewSessionShort = flagDef(flag.Bool, "new", "n", "save any existing session and start a new one", false)
	flags.ListSessions, flags.ListSessionsShort = flagDef(flag.Bool, "list", "l", "list all sessions by id", false)
	flags.RestoreSession, flags.RestoreSessionShort = flagDef(flag.Int, "restore", "r", "the session id to restore", 0)
	flags.DeleteSession, flags.DeleteSessionShort = flagDef(flag.Int, "delete", "d", "the session id to delete", 0)

	flags.CustomURL = flag.String("url", "", "a custom url to use for the gemini api. by default the vertex-ai (gcp) or generative-language-api (ai-studio) canonical urls are used depending on whether "+
		"an access-token is specified or not. where no access-token is specified, the generative-language-api form is used and the GEMINI_API_KEY envar is queried for the api-key to include in its querystring. where an "+
		"access-token is specified, the vertex-ai form is used and both -gcp-project and -gcs-bucket arguments must be also specified. the following placeholders are supported in custom urls and will be populated "+
		"where specified and appropriate: {model}, {api-key}, {gcp-project}")

	flags.CustomUploadURL = flag.String("upload-url", "", "a custom url to use for file uploads. by default the cloud storage (gcp) or generative-language-api (ai-studio) canonical urls are used depending on whether "+
		"an access-token is specified or not. where no access-token is specified, the generative-language-api form is used and the GEMINI_API_KEY envar is queried for the api-key to include in its querystring. where an "+
		"access-token is specified, the cloud storage form is used and both -gcp-project and -gcs-bucket arguments must be also specified. the following placeholders are supported in custom urls and will be populated "+
		"where specified and appropriate: {api-key}, {gcs-bucket}, {file-name}")

	flags.VertexAccessToken, flags.VertexAccessTokenShort = flagDef(flag.String, "vertex-access-token", "a", "the access token to present to the vertex-ai (gcp) gemini api endpoint. specifying a vertex-access-token will cause the "+
		"vertex-ai (gcp) canonical endpoint to be used (unless a custom url is provided)", "")

	flags.GCPProject, flags.GCPProjectShort = flagDef(flag.String, "gcp-project", "p", "the gcp project to include in the gemini api url. specifying a gcp-project will cause the vertex-ai (gcp) canonical endpoint to be "+
		"used (unless a custom url is provided)", "")

	flags.GCSBucket, flags.GCSBucketShort = flagDef(flag.String, "gcs-bucket", "b", "the cloud storage (gcp) bucket to upload files to when using the gemini api via a vertex-ai (gcp) endpoint", "")

	flags.SchemaDefinition = flag.String("schema", "", "a schema that defines the required response format. either in the form 'field1:field1-type:field1-description|field2:field2-type:field2-description|...n' or "+
		"as a json-form open-api schema. grounding with search must be disabled to use a schema")

	flags.DeleteAllSessions = flag.Bool("delete-all", false, "delete all session data")
	flags.DisableGrounding = flag.Bool("no-grounding", false, "disable grounding with search")
	flags.Debug = flag.Bool("debug", false, "enable debug output")
	flags.Stats = flag.Bool("stats", false, "print count of tokens used")
	flags.AppDir = flag.String("app-dir", path.Join(homeDir, "."+app), fmt.Sprintf("location of the %v app directory", app))
	flags.CustomModel = flag.String("model", "", "the specific model to use")
	flags.FlashModel = flag.Bool("flash", false, fmt.Sprintf("use the cheaper %v model", flashModel))
	flags.MaxTokens = flag.Int("max-tokens", 10000, "the maximum number of tokens to allow in a response")
	flags.Temperature = flag.Float64("temperature", 0.2, "the temperature setting for the model")
	flags.TopP = flag.Float64("top-p", 0.2, "the top-p setting for the model")

	flags.SystemPrompt = flag.String("system-prompt",
		fmt.Sprintf("You are a command line utility named '%v' running in a terminal on the OS '%v'. Factor that into the format and content of your responses and always ensure they are concise and "+
			"easily rendered in such a terminal. You do not use complex markdown syntax in your responses as this is not rendered well in terminal output. You do use clear, plain text formatting that can be easily read "+
			"by a human; such as using dashes for list delimiters. You always ensure that, to the extent that you are reasonably able, that your answers are factually correct and you take caution regarding hallucinations. "+
			"You only answer the specific question given and do not proactively include additional information that is not directly relevant to that question. ", app, runtime.GOOS),
		"the system prompt to use")

	flags.UseCase = flag.String("use-case", "", "free text information to include in the system prompt about the user or use-case, such as a role or location. "+
		"for example 'you are running in a ci pipeline used to verify code quality' or 'you are assisting a go/linux software engineer based in staffordshire'")

	flag.Parse()

	return flags
}

// flagDef is a helper function to define a flag and its shortform
func flagDef[T any](flagFunc func(string, T, string) *T, name, shortform, desc string, val T) (*T, *T) {
	f := flagFunc(name, val, desc)
	fs := flagFunc(shortform, val, "shortform of -"+name)

	return f, fs
}
