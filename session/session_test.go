package session_test

import (
	"os"
	"testing"

	"github.com/comradequinn/gen/gemini"
	"github.com/comradequinn/gen/session"
)

func TestSession(t *testing.T) {
	testDir := "./test"
	os.RemoveAll(testDir)

	defer os.RemoveAll(testDir)

	writeSession := func(prompt, response string) {
		if err := session.Write(testDir,
			gemini.Transaction{
				Input:  gemini.Input{Text: prompt},
				Output: gemini.Output{Text: response},
			}); err != nil {
			t.Fatalf("expected no error writing session. got %v", err)
		}
	}

	writeSession("test-prompt-1", "test-response-1")
	writeSession("test-prompt-2", "test-response-2")
	writeSession("test-prompt-3", "test-response-3")

	actualsession, err := session.Read(testDir)

	if err != nil {
		t.Fatalf("expected no error reading session. got %v", err)
	}

	assertInt := func(actual, expected int, desc string) {
		if actual != expected {
			t.Fatalf("expected %v to be %v. got %v", desc, expected, actual)
		}
	}

	assertInt(len(actualsession), 3, "session count")

	assertString := func(actual, expected, message string) {
		if actual != expected {
			t.Fatalf("expected %v. got %v. %v", expected, actual, message)
		}
	}

	assertString(actualsession[0].Input.Text, "test-prompt-1", "first prompt")
	assertString(actualsession[0].Output.Text, "test-response-1", "first response")
	assertString(actualsession[1].Input.Text, "test-prompt-2", "second prompt")
	assertString(actualsession[1].Output.Text, "test-response-2", "second response")
	assertString(actualsession[2].Input.Text, "test-prompt-3", "third prompt")
	assertString(actualsession[2].Output.Text, "test-response-3", "third response")

	if err := session.Stash(testDir); err != nil {
		t.Fatalf("expected no error stashing session. got %v", err)
	}

	if actualsession, err = session.Read(testDir); err != nil {
		t.Fatalf("expected no error reading session. got %v", err)
	}

	assertInt(len(actualsession), 0, "session count")

	writeSession("test-prompt-4", "test-response-4")

	records, err := session.List(testDir)

	if err != nil {
		t.Fatalf("expected no error listing sessions. got %v", err)
	}

	assertInt(len(records), 2, "record count")

	if !records[1].Active {
		t.Fatalf("expected latest session to be active. got %+v", records)
	}

	if err := session.Restore(testDir, 1); err != nil {
		t.Fatalf("expected no error restoring session. got %v", err)
	}

	records, err = session.List(testDir)

	if err != nil {
		t.Fatalf("expected no error listing sessions. got %v", err)
	}

	if !records[0].Active {
		t.Fatalf("expected restored session to be active. got %+v", records)
	}

	if err := session.Delete(testDir, 2); err != nil {
		t.Fatalf("expected no error deleting session. got %v", err)
	}

	records, err = session.List(testDir)

	if err != nil {
		t.Fatalf("expected no error listing sessions. got %v", err)
	}

	assertInt(len(records), 1, "record count")

	if err := session.DeleteAll(testDir); err != nil {
		t.Fatalf("expected no error deleting all sessions. got %v", err)
	}

	records, err = session.List(testDir)

	if err != nil {
		t.Fatalf("expected no error listing sessions. got %v", err)
	}

	assertInt(len(records), 0, "record count")
}
