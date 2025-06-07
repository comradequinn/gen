package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

var (
	Write = func(format string, a ...any) {
		if !strings.HasSuffix(format, "\n") {
			format += "\n"
		}
		fmt.Printf(format, a...)
	}
	WriteInfo            = func(format string, a ...any) { Write(fmt.Sprintf("\x1b[90m%v\x1b[0m", format), a...) }
	WriteError           = func(format string, a ...any) { Write(fmt.Sprintf("\x1b[31m%v\x1b[0m", format), a...) }
	WriteRaw             = func(format string, a ...any) { fmt.Printf(format, a...) }
	Reader     io.Reader = os.Stdin
)

func spin() (stopFunc func()) {
	taskDone, spinDone := make(chan struct{}), make(chan struct{})

	go func() {
		spinner, i := []rune{'|', '/', '-', '\\'}, 0

		for {
			select {
			case <-taskDone:
				WriteRaw("\r")
				spinDone <- struct{}{}
				return
			default:
				i++
				WriteRaw("\r%c", spinner[i%len(spinner)])
				<-time.After(150 * time.Millisecond)
			}
		}
	}()

	return func() {
		taskDone <- struct{}{}
		<-spinDone
	}
}
