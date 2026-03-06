//go:build windows

package main

import "io/fs"

func readStatFields(_ fs.FileInfo, _ string) (statFields, error) {
	return statFields{nlink: 1, uid: 0, gid: 0}, nil
}
