package cli

import (
	"fmt"
	"strings"

	"github.com/comradequinn/gen/session"
)

// ListSessions displays the current and any saved sessions
func ListSessions(records []session.Record) {
	for i, r := range records {
		labelPrefix := "  "

		if r.Active {
			labelPrefix = "* "
		}

		Write(fmt.Sprintf("%v #%v (%v): %v\n", labelPrefix, i+1, r.TimeStamp.Format("January 02 2006"), strings.ToLower(r.Summary)))
	}
}
