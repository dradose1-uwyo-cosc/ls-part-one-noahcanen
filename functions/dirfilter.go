package functions

import "os"

func dirFilter(entries []os.DirEntry) []os.DirEntry {
	out := entries[:0]
	for _, e := range entries {
		name := e.Name()
		if len(name) > 0 && name[0] == '.' {
			continue
		}
		out = append(out, e)
	}
	return out
}
