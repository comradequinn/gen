package main

import (
	"flag"
	"os"
	"strings"

	"github.com/comradequinn/gen/cli"
	"github.com/comradequinn/gen/gemini"
	"github.com/comradequinn/gen/log"
	"github.com/comradequinn/gen/schema"
	"github.com/comradequinn/gen/session"
)

const (
	app = "gen"
)

var (
	commit = "dev-build"
	tag    = "v0.0.0-dev.0"
)

func main() {
	defer func() {
		err := recover()
		log.FatalfIf(err != nil, "process terminated due to panic. %v", err)
	}()

	args := cli.ReadArgs(os.Getenv("HOME"), app, gemini.Models.Pro)

	log.Init(*args.Debug || *args.DebugShort, cli.WriteError)

	apiCustomURL, uploadCustomURL := *args.CustomURL, *args.CustomUploadURL
	gcpProjectID, apiCredential, gcsBucketName := *args.GCPProject+*args.GCPProjectShort, *args.VertexAccessToken+*args.VertexAccessTokenShort, *args.GCSBucket+*args.GCSBucketShort
	commandExecutionEnabled, commandApprovalEnabled := *args.CommandExecution || *args.CommandExecutionShort, *args.CommandApproval || *args.CommandApprovalShort
	scriptMode := *args.Script || *args.ScriptShort

	log.FatalfIf(scriptMode && commandApprovalEnabled, "command approval cannot be enabled in script mode")

	if apiCredential == "" {
		apiCredential = os.Getenv("GEMINI_API_KEY")
	}

	{ // non-prompt commands
		switch {
		case *args.Version:
			cli.Write("%v %v %v (pro-model: %v, flash-model: %v)\n", app, tag, commit, gemini.Models.Pro, gemini.Models.Flash)
			os.Exit(0)
		case *args.NewSession || *args.NewSessionShort:
			session.Stash(*args.AppDir)
		case *args.RestoreSession > 0 || *args.RestoreSessionShort > 0:
			err := session.Restore(*args.AppDir, *args.RestoreSession+*args.RestoreSessionShort)
			log.FatalfIf(err != nil, "unable to restore session. %v", err)
			os.Exit(0)
		case *args.DeleteSession > 0 || *args.DeleteSessionShort > 0:
			err := session.Delete(*args.AppDir, *args.DeleteSession+*args.DeleteSessionShort)
			log.FatalfIf(err != nil, "unable to delete session. %v", err)
			os.Exit(0)
		case *args.DeleteAllSessions:
			err := session.DeleteAll(*args.AppDir)
			log.FatalfIf(err != nil, "unable to delete sessions. %v", err)
			os.Exit(0)
		case *args.ListSessions || *args.ListSessionsShort:
			records, err := session.List(*args.AppDir)
			log.FatalfIf(err != nil, "unable to list history. %v", err)
			cli.ListSessions(records)
			os.Exit(0)
		}
	}

	log.FatalfIf(len(flag.Args()) != 1, "a single prompt is required")

	promptText := flag.Arg(0)

	model := gemini.Models.Flash
	switch {
	case *args.CustomModel != "":
		model = *args.CustomModel
	case *args.ProModel:
		model = gemini.Models.Pro
	}

	schema, err := schema.Build(*args.SchemaDefinition + *args.SchemaDefinitionShort)
	log.FatalfIf(err != nil, "invalid schema definition. %v", err)

	files := []string{}
	{
		if filePattern := *args.Files + *args.FilesShort; filePattern != "" {
			files = strings.Split(filePattern, ",")
			for i := range files {
				files[i] = strings.TrimSpace(files[i])
			}
		}
	}

	cli.Generate(gemini.Config{
		GeminiURL:        apiCustomURL,
		Credential:       apiCredential,
		GCPProject:       gcpProjectID,
		GCSBucket:        gcsBucketName,
		Model:            model,
		FileStorageURL:   uploadCustomURL,
		SystemPrompt:     *args.SystemPrompt,
		MaxTokens:        *args.MaxTokens,
		Temperature:      *args.Temperature,
		TopP:             *args.TopP,
		Grounding:        !*args.DisableGrounding,
		UseCase:          *args.UseCase,
		CommandExecution: commandExecutionEnabled,
		CommandApproval:  commandApprovalEnabled,
	}, args, scriptMode, promptText, schema, files)
}
