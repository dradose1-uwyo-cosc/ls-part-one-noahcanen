package main

import (
	"os"

	"lspartone/functions"
)

func main() {
	args := os.Args[1:]
	useColor := functions.IsTerminal(os.Stdout)
	functions.SimpleLS(os.Stdout, args, useColor)
}
