package editorcmd

import (
	"os"
	"os/exec"
	"runtime"
)

// Detect detects a text editor command.
// Returns an empty string when no editor is found.
func Detect() string {
	var candidates = []string{
		os.Getenv("VISUAL"),
		os.Getenv("EDITOR"),
		"editor",
		"vim",
		"vi",
		"emacs",
	}
	if runtime.GOOS == "windows" {
		candidates = append(candidates, "notepad.exe")
	}
	for _, f := range candidates {
		if f == "" {
			continue
		}
		x, err := exec.LookPath(f)
		if err == nil {
			return x
		}
	}
	return ""
}
