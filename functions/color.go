package functions

import (
	"fmt"
	"io"
)

const (
	ansiReset = "\x1b[0m"
	ansiBlue  = "\x1b[34m"
	ansiGreen = "\x1b[32m"
)

func ColorBlue(w io.Writer, s string) {
	fmt.Fprintf(w, "%s%s%s\n", ansiBlue, s, ansiReset)
}

func ColorGreen(w io.Writer, s string) {
	fmt.Fprintf(w, "%s%s%s\n", ansiGreen, s, ansiReset)
}
