//go:build !windows

package main

import (
	"fmt"
	"io/fs"
	"syscall"
)

func readStatFields(info fs.FileInfo, fullPath string) (statFields, error) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return statFields{}, fmt.Errorf("missing syscall.Stat_t for %s", fullPath)
	}
	return statFields{
		nlink: uint64(stat.Nlink),
		uid:   uint64(stat.Uid),
		gid:   uint64(stat.Gid),
	}, nil
}
