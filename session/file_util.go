package session

import (
	"fmt"
	"math/rand/v2"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

func sessionDir(appDir string) (string, error) {
	if appDir == "" {
		panic("session directory location not set")
	}

	var sessionDir = path.Join(appDir, "session")

	if _, err := os.Stat(sessionDir); os.IsNotExist(err) {
		if err := os.MkdirAll(sessionDir, 0755); err != nil {
			return "", fmt.Errorf("unable to create session directory. %w", err)
		}
	}

	return sessionDir, nil
}

func activeSessionFilePath(appDir string) (string, bool, error) {
	sessionDir, err := sessionDir(appDir)
	if err != nil {
		return "", false, err
	}

	files, err := os.ReadDir(sessionDir)

	if err != nil {
		return "", false, fmt.Errorf("unable to read session directory. %w", err)
	}

	for _, f := range files {
		if !f.Type().IsRegular() {
			continue
		}

		if strings.HasSuffix(f.Name(), ActiveSessionFileSuffix) {
			return path.Join(sessionDir, f.Name()), true, nil
		}
	}

	return "", false, nil
}

func openActiveSessionFile(appDir string, flag int) (*os.File, error) {
	sessionDir, err := sessionDir(appDir)
	if err != nil {
		return nil, err
	}

	sessionFilePath, exists, err := activeSessionFilePath(appDir)
	if err != nil {
		return nil, err
	}

	if exists {
		sessionFile, err := os.OpenFile(sessionFilePath, flag, 0600)
		if err != nil {
			return nil, fmt.Errorf("unable to open session file. %w", err)
		}
		return sessionFile, nil
	}

	sessionFile, err := os.OpenFile(path.Join(sessionDir, strconv.FormatInt(time.Now().UnixNano(), 10)+"_"+strconv.Itoa(rand.Int())+ActiveSessionFileSuffix), flag|os.O_CREATE, 0600)
	if err != nil {
		return nil, fmt.Errorf("unable to open session file. %w", err)
	}

	return sessionFile, nil
}
