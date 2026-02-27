package functions

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

func SimpleLS(w io.Writer, args []string, useColor bool) {
	if len(args) == 0 {
		listDirectory(w, ".", useColor)
		return
	}

	var fileTargets []string
	var dirTargets []string

	for _, arg := range args {
		info, err := os.Lstat(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gols: cannot access '%s': %v\n", arg, err)
			continue
		}
		if info.IsDir() {
			dirTargets = append(dirTargets, arg)
		} else {
			fileTargets = append(fileTargets, arg)
		}
	}

	sort.Strings(fileTargets)
	sort.Strings(dirTargets)

	for _, f := range fileTargets {
		name := filepath.Base(f)
		info, err := os.Lstat(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gols: cannot access '%s': %v\n", f, err)
			continue
		}
		writeName(w, name, info.Mode(), info.IsDir(), useColor)
	}

	if len(fileTargets) > 0 && len(dirTargets) > 0 {
		fmt.Fprintln(w)
	}

	showHeaders := (len(dirTargets) > 1) || (len(fileTargets) > 0 && len(dirTargets) > 0)
	for index, d := range dirTargets {
		if showHeaders {
			fmt.Fprintf(w, "%s:\n", d)
		}
		listDirectory(w, d, useColor)
		if index != len(dirTargets)-1 {
			fmt.Fprintln(w)
		}
	}
}

func listDirectory(w io.Writer, dir string, useColor bool) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gols: cannot access '%s': %v\n", dir, err)
		return
	}

	entries = dirFilter(entries)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, de := range entries {
		full := filepath.Join(dir, de.Name())
		fi, err := os.Lstat(full)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gols: cannot access '%s': %v\n", full, err)
			continue
		}
		writeName(w, de.Name(), fi.Mode(), de.IsDir(), useColor)
	}
}

func writeName(w io.Writer, name string, mode os.FileMode, isDir bool, useColor bool) {
	if useColor {
		switch {
		case isDir:
			ColorBlue(w, name)
			return
		case mode.IsRegular() && (mode&0o111) != 0:
			ColorGreen(w, name)
			return
		}
	}
	fmt.Fprintln(w, name)
}
