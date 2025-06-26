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

	log.Init(args.Debug(), cli.WriteError)

	log.FatalfIf(args.Quiet() && args.ExecutionApproval(), "command approval cannot be enabled in quiet mode")

	apiCredential := args.AccessToken()

	if apiCredential == "" {
		apiCredential = os.Getenv("GEMINI_API_KEY")
	}

	{ // non-prompt commands
		switch {
		case *args.Version:
			cli.Write("%v %v %v (pro-model: %v, flash-model: %v)\n", app, tag, commit, gemini.Models.Pro, gemini.Models.Flash)
			os.Exit(0)
		case args.RestoreSession() > 0:
			err := session.Restore(*args.AppDir, args.RestoreSession())
			log.FatalfIf(err != nil, "unable to restore session. %v", err)
			os.Exit(0)
		case args.DeleteSession() > 0:
			err := session.Delete(*args.AppDir, args.DeleteSession())
			log.FatalfIf(err != nil, "unable to delete session. %v", err)
			os.Exit(0)
		case *args.DeleteAllSessions:
			err := session.DeleteAll(*args.AppDir)
			log.FatalfIf(err != nil, "unable to delete sessions. %v", err)
			os.Exit(0)
		case args.ListSessions():
			records, err := session.List(*args.AppDir)
			log.FatalfIf(err != nil, "unable to list history. %v", err)
			cli.ListSessions(records)
			os.Exit(0)
		}
	}

	if !args.ContinueSession() {
		session.Stash(*args.AppDir)
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

	schema, err := schema.Build(args.SchemaDefinition())
	log.FatalfIf(err != nil, "invalid schema definition. %v", err)

	filePaths := []string{}
	{
		if filePattern := args.FilePaths(); filePattern != "" {
			filePaths = strings.Split(filePattern, ",")
			for i := range filePaths {
				filePaths[i] = strings.TrimSpace(filePaths[i])
			}
		}
	}

	cli.Generate(gemini.Config{
		GeminiURL:         *args.CustomURL,
		Credential:        apiCredential,
		GCPProject:        args.GCPProject(),
		GCSBucket:         args.GCSBucket(),
		Model:             model,
		FileStorageURL:    *args.CustomUploadURL,
		SystemPrompt:      *args.SystemPrompt,
		MaxTokens:         *args.MaxTokens,
		Temperature:       *args.Temperature,
		TopP:              *args.TopP,
		Grounding:         !*args.DisableGrounding,
		UseCase:           *args.UseCase,
		ExecutionEnabled:  args.ExecutionEnabled(),
		ExecutionApproval: args.ExecutionApproval(),
	}, args, args.Quiet(), promptText, schema, filePaths)
}
