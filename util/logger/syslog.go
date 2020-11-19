package logger

import (
	"strings"

	slog "github.com/RackSec/srslog"
)

func Err(w *slog.Writer, e error, t string) {
	if w != nil {
		w.Err(t + " " + e.Error())
	}
}

func Verbose(w *slog.Writer, s []string) {
	if w != nil {
		w.Info(strings.Join(s, " "))
	}
}
