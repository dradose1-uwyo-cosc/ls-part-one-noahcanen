package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	colorReset = "\033[0m"
	colorBlue  = "\033[34m"
	colorGreen = "\033[32m"
)

type options struct {
	all       bool
	long      bool
	numeric   bool
	human     bool
	recursive bool
	colorize  bool
}

type entry struct {
	name         string
	fullPath     string
	info         fs.FileInfo
	perm         string
	links        string
	owner        string
	group        string
	size         string
	modTime      string
	displayName  string
	recurseTo    bool
	isDirectory  bool
	isExecutable bool
}

type widths struct {
	links int
	owner int
	group int
	size  int
}

type statFields struct {
	nlink uint64
	uid   uint64
	gid   uint64
}

func main() {
	opts := parseFlags()
	args := flag.Args()
	if len(args) == 0 {
		args = []string{"."}
	}

	exitCode := 0
	multipleTargets := len(args) > 1

	for i, target := range args {
		if i > 0 {
			fmt.Println()
		}
		if err := listTarget(target, opts, multipleTargets); err != nil {
			exitCode = 1
			fmt.Fprintf(os.Stderr, "gols: %s: %v\n", target, err)
		}
	}

	if exitCode != 0 {
		os.Exit(exitCode)
	}
}

func parseFlags() options {
	all := flag.Bool("a", false, "include entries starting with dot")
	long := flag.Bool("l", false, "long listing format")
	numeric := flag.Bool("n", false, "numeric uid/gid")
	human := flag.Bool("h", false, "human-readable sizes")
	recursive := flag.Bool("R", false, "recursive listing")

	flag.Parse()

	outInfo, err := os.Stdout.Stat()
	colorize := err == nil && (outInfo.Mode()&os.ModeCharDevice) != 0

	finalLong := *long || *numeric

	return options{
		all:       *all,
		long:      finalLong,
		numeric:   *numeric,
		human:     *human,
		recursive: *recursive,
		colorize:  colorize,
	}
}

func listTarget(target string, opts options, multipleTargets bool) error {
	info, err := os.Lstat(target)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		entries, err := gatherFileEntry(target, opts)
		if err != nil {
			return err
		}
		printEntries(entries, opts)
		return nil
	}

	cleanTarget := target
	if cleanTarget == "" {
		cleanTarget = "."
	}

	if opts.recursive {
		return walkDirectory(cleanTarget, opts)
	}

	if multipleTargets {
		fmt.Printf("%s:\n", cleanTarget)
	}

	entries, err := gatherDirectoryEntries(cleanTarget, opts)
	if err != nil {
		return err
	}
	printEntries(entries, opts)
	return nil
}

func walkDirectory(dir string, opts options) error {
	fmt.Printf("%s:\n", dir)
	entries, err := gatherDirectoryEntries(dir, opts)
	if err != nil {
		return err
	}
	printEntries(entries, opts)

	for _, e := range entries {
		if !e.recurseTo {
			continue
		}
		fmt.Println()
		if err := walkDirectory(e.fullPath, opts); err != nil {
			fmt.Fprintf(os.Stderr, "gols: %s: %v\n", e.fullPath, err)
		}
	}
	return nil
}

func gatherFileEntry(path string, opts options) ([]entry, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}

	e, err := buildEntry(filepath.Base(path), path, info, opts, false)
	if err != nil {
		return nil, err
	}
	return []entry{e}, nil
}

func gatherDirectoryEntries(dir string, opts options) ([]entry, error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	allNames := make([]string, 0, len(dirEntries)+2)
	if opts.all {
		allNames = append(allNames, ".", "..")
	}
	for _, de := range dirEntries {
		name := de.Name()
		if !opts.all && strings.HasPrefix(name, ".") {
			continue
		}
		allNames = append(allNames, name)
	}

	sort.Strings(allNames)

	entries := make([]entry, 0, len(allNames))
	for _, name := range allNames {
		fullPath := filepath.Join(dir, name)
		info, err := os.Lstat(fullPath)
		if err != nil {
			return nil, err
		}
		e, err := buildEntry(name, fullPath, info, opts, true)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func buildEntry(name, fullPath string, info fs.FileInfo, opts options, fromDirectory bool) (entry, error) {
	stat, err := readStatFields(info, fullPath)
	if err != nil {
		return entry{}, err
	}

	perm := permissionString(info.Mode())
	links := strconv.FormatUint(stat.nlink, 10)

	uid := strconv.FormatUint(stat.uid, 10)
	gid := strconv.FormatUint(stat.gid, 10)

	owner := uid
	group := gid
	if !opts.numeric {
		if u, err := user.LookupId(uid); err == nil {
			owner = u.Username
		}
		if g, err := user.LookupGroupId(gid); err == nil {
			group = g.Name
		}
	}

	sizeVal := info.Size()
	size := strconv.FormatInt(sizeVal, 10)
	if opts.human {
		size = humanSize(sizeVal)
	}

	modTime := formatModTime(info.ModTime(), time.Now())

	isDir := info.IsDir()
	isSymlink := info.Mode()&os.ModeSymlink != 0
	isExec := (info.Mode().Perm()&0o111) != 0 && !isDir

	displayName := colorName(name, isDir, isExec, opts.colorize)

	recurseTo := fromDirectory && isDir && !isSymlink && name != "." && name != ".."

	return entry{
		name:         name,
		fullPath:     fullPath,
		info:         info,
		perm:         perm,
		links:        links,
		owner:        owner,
		group:        group,
		size:         size,
		modTime:      modTime,
		displayName:  displayName,
		recurseTo:    recurseTo,
		isDirectory:  isDir,
		isExecutable: isExec,
	}, nil
}

func permissionString(mode os.FileMode) string {
	var b [10]byte

	switch {
	case mode&os.ModeSymlink != 0:
		b[0] = 'l'
	case mode.IsDir():
		b[0] = 'd'
	default:
		b[0] = '-'
	}

	permBits := []struct {
		bit os.FileMode
		ch  byte
	}{
		{0o400, 'r'}, {0o200, 'w'}, {0o100, 'x'},
		{0o040, 'r'}, {0o020, 'w'}, {0o010, 'x'},
		{0o004, 'r'}, {0o002, 'w'}, {0o001, 'x'},
	}

	for i, p := range permBits {
		if mode&p.bit != 0 {
			b[i+1] = p.ch
		} else {
			b[i+1] = '-'
		}
	}

	return string(b[:])
}

func formatModTime(mod, now time.Time) string {
	localMod := mod.Local()
	localNow := now.Local()
	cutoff := localNow.AddDate(0, 0, -180)

	if !localMod.Before(cutoff) && !localMod.After(localNow) {
		return localMod.Format("Jan _2 3:04")
	}
	return localMod.Format("Jan _2  2006")
}

func humanSize(size int64) string {
	if size < 1024 {
		return strconv.FormatInt(size, 10)
	}

	units := []string{"K", "M", "G", "T"}
	val := float64(size)
	unit := ""
	for _, u := range units {
		val /= 1024.0
		unit = u
		if val < 1024.0 {
			break
		}
	}

	formatted := fmt.Sprintf("%.1f", val)
	formatted = strings.TrimSuffix(formatted, ".0")
	return formatted + unit
}

func colorName(name string, isDir, isExec, enabled bool) string {
	if !enabled {
		return name
	}
	if isDir {
		return colorBlue + name + colorReset
	}
	if isExec {
		return colorGreen + name + colorReset
	}
	return name
}

func computeWidths(entries []entry) widths {
	w := widths{}
	for _, e := range entries {
		if len(e.links) > w.links {
			w.links = len(e.links)
		}
		if len(e.owner) > w.owner {
			w.owner = len(e.owner)
		}
		if len(e.group) > w.group {
			w.group = len(e.group)
		}
		if len(e.size) > w.size {
			w.size = len(e.size)
		}
	}
	return w
}

func printEntries(entries []entry, opts options) {
	if !opts.long {
		for _, e := range entries {
			fmt.Println(e.displayName)
		}
		return
	}

	w := computeWidths(entries)
	for _, e := range entries {
		fmt.Printf("%-10s %*s %-*s %-*s %*s %s %s\n",
			e.perm,
			w.links, e.links,
			w.owner, e.owner,
			w.group, e.group,
			w.size, e.size,
			e.modTime,
			e.displayName,
		)
	}
}
