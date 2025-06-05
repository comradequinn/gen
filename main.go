package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/comradequinn/gen/cli"
	"github.com/comradequinn/gen/gemini"
	"github.com/comradequinn/gen/schema"
	"github.com/comradequinn/gen/session"
)

const (
	app = "gen"
)

var (
	commit = "dev"
	tag    = "none"
)

func main() {
	checkFatalf := func(condition bool, format string, v ...any) {
		if !condition {
			return
		}
		fmt.Printf(format+"\n", v...)
		os.Exit(1)
	}

	args := cli.ReadArgs(os.Getenv("HOME"), app, gemini.Models.Flash)

	logLevel := slog.LevelInfo

	if *args.Debug {
		logLevel = slog.LevelDebug
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})))

	apiURL, uploadURL := *args.CustomURL, *args.CustomUploadURL
	gcpProjectID, apiCredential, gcsBucketName := *args.GCPProject+*args.GCPProjectShort, *args.VertexAccessToken+*args.VertexAccessTokenShort, *args.GCSBucket+*args.GCSBucketShort
	platform, scriptMode := gemini.PlatformGenerativeLanguage, *args.Script || *args.ScriptShort

	if gcpProjectID != "" || apiCredential != "" || gcsBucketName != "" {
		checkFatalf(gcpProjectID == "" || apiCredential == "" || gcsBucketName == "", "to use the gemini api via vertex-ai a gcp-project, gcs-bucket and vertex-access-token must be provided")
		platform = gemini.PlatformVertex
	}

	switch platform {
	case gemini.PlatformGenerativeLanguage:
		if apiURL == "" {
			apiURL = "https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent?key={api-key}"
		}
		if uploadURL == "" {
			uploadURL = "https://generativelanguage.googleapis.com/upload/v1beta/files?key={api-key}"
		}
		apiCredential = os.Getenv("GEMINI_API_KEY")
	case gemini.PlatformVertex:
		if apiURL == "" {
			apiURL = "https://aiplatform.googleapis.com/v1/projects/{gcp-project}/locations/global/publishers/google/models/{model}:generateContent"
		}
		if uploadURL == "" {
			uploadURL = "https://storage.googleapis.com/upload/storage/v1/b/{gcs-bucket}/o?uploadType=media&name={file-name}"
		}
	}

	model := gemini.Models.Pro
	switch {
	case *args.CustomModel != "":
		model = *args.CustomModel
	case *args.FlashModel:
		model = gemini.Models.Flash
	}

	formatURL := func(u string) string {
		return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(u,
			"{model}", model),
			"{gcp-project}", gcpProjectID),
			"{gcs-bucket}", gcsBucketName)
	}

	apiURL, uploadURL = formatURL(apiURL), formatURL(uploadURL)

	{ // non-prompt commands
		switch {
		case *args.Version || *args.VersionShort:
			fmt.Printf("%v %v %v (pro-model: %v, flash-model: %v)\n", app, tag, commit, gemini.Models.Pro, gemini.Models.Flash)
			os.Exit(0)
		case *args.NewSession || *args.NewSessionShort:
			session.Stash(*args.AppDir)
		case *args.RestoreSession > 0 || *args.RestoreSessionShort > 0:
			err := session.Restore(*args.AppDir, *args.RestoreSession+*args.RestoreSessionShort)
			checkFatalf(err != nil, "unable to restore session. %v", err)
			os.Exit(0)
		case *args.DeleteSession > 0 || *args.DeleteSessionShort > 0:
			err := session.Delete(*args.AppDir, *args.DeleteSession+*args.DeleteSessionShort)
			checkFatalf(err != nil, "unable to delete session. %v", err)
			os.Exit(0)
		case *args.DeleteAllSessions:
			err := session.DeleteAll(*args.AppDir)
			checkFatalf(err != nil, "unable to delete sessions. %v", err)
			os.Exit(0)
		case *args.ListSessions || *args.ListSessionsShort:
			records, err := session.List(*args.AppDir)
			checkFatalf(err != nil, "unable to list history. %v", err)
			cli.ListSessions(records)
			os.Exit(0)
		}
	}

	checkFatalf(len(flag.Args()) != 1, "a single prompt is required")
	prompt := flag.Arg(0)

	var stopSpinner = func() {}
	{
		if !scriptMode {
			stopSpinner = cli.Spin()
		}
	}

	schema, err := schema.Build(*args.SchemaDefinition)
	checkFatalf(err != nil, "invalid schema definition. %v", err)

	messages, err := session.Read(*args.AppDir)
	checkFatalf(err != nil, "unable to read history. %v", err)

	files := []string{}
	{
		if filePattern := *args.Files + *args.FilesShort; filePattern != "" {
			files = strings.Split(filePattern, ",")
			for i := range files {
				files[i] = strings.TrimSpace(files[i])
			}
		}
	}

	rs, err := gemini.Generate(
		gemini.Config{
			Platform:       platform,
			GeminiURL:      apiURL,
			Credential:     apiCredential,
			FileStorageURL: uploadURL,
			SystemPrompt:   *args.SystemPrompt,
			MaxTokens:      *args.MaxTokens,
			Temperature:    *args.Temperature,
			TopP:           *args.TopP,
			Grounding:      !*args.DisableGrounding,
			UseCase:        *args.UseCase,
		},
		gemini.Prompt{
			Text:    prompt,
			Files:   files,
			History: messages,
			Schema:  schema,
		}, slog.Debug)

	checkFatalf(err != nil, "error with gemini api. %v", err)

	checkFatalf(session.Write(*args.AppDir, session.Entry{
		Prompt:   prompt,
		Response: rs.Text,
		Files:    rs.Files,
	}) != nil, "unable to update session. %v", err)

	stopSpinner()

	fmt.Printf("%v\n\n", rs.Text)

	if *args.Stats {
		_ = json.NewEncoder(os.Stderr).Encode(map[string]map[string]string{
			"stats": {
				"systemPromptBytes": fmt.Sprintf("%v", len(*args.SystemPrompt)),
				"promptBytes":       fmt.Sprintf("%v", len(prompt)),
				"responseBytes":     fmt.Sprintf("%v", len(rs.Text)),
				"tokens":            fmt.Sprintf("%v", rs.Tokens),
				"files":             fmt.Sprintf("%v", len(rs.Files)),
			},
		})
	}
}
