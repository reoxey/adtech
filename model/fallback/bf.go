package fallback

import (
	"net/http"
	"strings"
)

// Fallback redirects invalid click traffic to pre-specified options
// Smartlink or individual campaign
type Fallback struct {
	w    http.ResponseWriter
	Stop string
	To   string
	Logs strings.Builder
}

// BF implements Fallback
type BF interface {
	Redirect(string)
}

var _ BF = (*Fallback) (nil)

// Init initialised Fallback
func Init(w http.ResponseWriter, st, to string, l strings.Builder) BF {
	return Fallback{w, st, to, l}
}

// Redirect click traffic if invalid
func (f Fallback) Redirect(m string) { //TODO
	if f.Stop == "1" {
		f.w.Write([]byte("No Fallback"))
	}

	if f.To == "SM" {
		f.smartLink(m)
		return
	}

	f.w.Write([]byte("Fallback: " + f.To + " " + f.Logs.String() + m))
}

func (f Fallback) smartLink(m string) { //TODO
	f.w.Write([]byte("Fallback: SM " + f.Logs.String() + m))
}