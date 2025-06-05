package session

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/comradequinn/gen/gemini"
)

type (
	Entry struct {
		Prompt   string
		Files    []gemini.FileReference
		Response string
	}
	Record struct {
		ID        int
		Name      string
		Summary   string
		TimeStamp time.Time
		Active    bool
	}
)

const ActiveSessionFileSuffix = ".active"

// Write adds the specified entry to the active session
func Write(appDir string, entry Entry) error {
	messages, err := Read(appDir)

	if err != nil {
		return err
	}

	f, err := openActiveSessionFile(appDir, os.O_WRONLY|os.O_TRUNC)

	if err != nil {
		return err
	}

	defer f.Close()

	jsonEncoder := json.NewEncoder(f)
	jsonEncoder.SetIndent("", "  ")

	if err := jsonEncoder.Encode(append(messages, gemini.Message{
		Role:  gemini.RoleUser,
		Text:  entry.Prompt,
		Files: entry.Files,
	}, gemini.Message{
		Role: gemini.RoleModel,
		Text: entry.Response,
	})); err != nil {
		return fmt.Errorf("unable to encode session file. %w", err)
	}

	return nil
}

// Read returns all messages in the active session
func Read(appDir string) ([]gemini.Message, error) {
	f, err := openActiveSessionFile(appDir, os.O_RDONLY)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	messages := []gemini.Message{}

	if err := json.NewDecoder(f).Decode(&messages); err != nil {
		if err == io.EOF {
			return []gemini.Message{}, nil
		}

		return nil, fmt.Errorf("unable to decode session file. %w", err)
	}

	return messages, nil
}

// List returns summary and meta data for all saved sessions and the active one
func List(appDir string) ([]Record, error) {
	sessionDir, err := sessionDir(appDir)

	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(sessionDir)

	if err != nil {
		return nil, fmt.Errorf("unable to read session directory. %v", err)
	}

	summarise := func(f string) (string, error) {
		sessionFile, err := os.Open(path.Join(sessionDir, f))

		if err != nil {
			return "", err
		}

		defer sessionFile.Close()

		messages := []gemini.Message{}

		if err := json.NewDecoder(sessionFile).Decode(&messages); err != nil && err != io.EOF {
			return "", fmt.Errorf("unable to decode session file. %v %v", sessionFile, err)
		}

		if len(messages) == 0 {
			return "[ no content ]", nil
		}

		const limit = 50

		if len(messages[0].Text) < limit {
			return messages[0].Text, nil
		}

		return messages[0].Text[:limit] + "...", nil
	}

	records := make([]Record, 0, len(files))

	for i, f := range files { // files are sorted by file name which is based on the unixnano time at the point of writing; which is the sort order required
		if !f.Type().IsRegular() {
			continue
		}

		summary, err := summarise(f.Name())

		if err != nil {
			return nil, fmt.Errorf("unable to summarise session file %v. %v", f.Name(), err)
		}

		info, err := f.Info()

		if err != nil {
			return nil, fmt.Errorf("unable to get timestamp for session file %v. %v", f.Name(), err)
		}

		records = append(records, Record{
			ID:        i + 1,
			Name:      f.Name(),
			Summary:   summary,
			TimeStamp: info.ModTime(),
			Active:    strings.HasSuffix(f.Name(), ActiveSessionFileSuffix),
		})
	}

	sort.SliceStable(records, func(i, j int) bool {
		return records[i].TimeStamp.Before(records[j].TimeStamp)
	})

	return records, nil
}

// Stash saves the current session and starts a new one
func Stash(appDir string) error {
	sessionFile, exists, err := activeSessionFilePath(appDir)

	if !exists || err != nil {
		return err
	}

	if err := os.Rename(sessionFile, strings.TrimSuffix(sessionFile, ActiveSessionFileSuffix)); err != nil {
		return fmt.Errorf("unable to rename existing active	session file. %w", err)
	}

	return nil
}

// Restore sets the specified stashed session as the active session
func Restore(appDir string, recordID int) error {
	records, err := List(appDir)

	if err != nil {
		return err
	}

	if recordID > len(records) {
		return fmt.Errorf("invalid record id %v", recordID)
	}

	record := records[recordID-1]

	if record.Active {
		return nil
	}

	sessionDir, err := sessionDir(appDir)
	if err != nil {
		return err
	}

	if err := Stash(appDir); err != nil {
		return err
	}

	if err := os.Rename(path.Join(sessionDir, record.Name), path.Join(sessionDir, record.Name+ActiveSessionFileSuffix)); err != nil {
		return fmt.Errorf("unable to restore session file. %w", err)
	}

	return nil
}

// Delete removes the specified session
func Delete(appDir string, recordID int) error {
	records, err := List(appDir)

	if err != nil {
		return err
	}

	if recordID > len(records) {
		return fmt.Errorf("invalid record id %v", recordID)
	}

	record := records[recordID-1]

	sessionDir, err := sessionDir(appDir)
	if err != nil {
		return err
	}

	if err := os.Remove(path.Join(sessionDir, record.Name)); err != nil {
		return fmt.Errorf("unable to delete session file. %w", err)
	}

	return nil
}

// DeleteAll removes all stashed sessions
func DeleteAll(appDir string) error {
	sessionDir, err := sessionDir(appDir)

	if err != nil {
		return err
	}

	if err := os.RemoveAll(sessionDir); err != nil {
		return fmt.Errorf("unable to delete all session data. %w", err)
	}

	return nil
}
