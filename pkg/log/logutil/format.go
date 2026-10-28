package logutil

import (
	"fmt"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
)

// FormatSourceFileAndLine shortens the supplied [runtime.Frame] filepath to
// just the last 3 file path parts (inclusive of the base filename) and returns
// a string containingn that shortened filepath and a line number from the
// [runtime.Frame].
//
// For example, given a [runtime.Frame] with a [runtime.Frame.File] value of
// /home/jaypipes/src/github.com/project/pkg/log/config.go and a
// [runtime.Frame.Line] value of 20, the returned value would be:
//
// pkg/log/config.go:20
func FormatSourceFileAndLine(f runtime.Frame) string {
	if f.File == "" {
		return ""
	}
	dirs := strings.Split(filepath.Dir(f.File), string(filepath.Separator))
	popDirs := min(len(dirs)-1, 2)
	slices.Reverse(dirs)
	dirs = dirs[0:popDirs]
	slices.Reverse(dirs)
	base := filepath.Base(f.File)
	dirs = append(dirs, base)
	fn := strings.Join(dirs, string(filepath.Separator))
	return fmt.Sprintf("%s:%d", fn, f.Line)
}
