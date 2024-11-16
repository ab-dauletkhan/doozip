package utils

import (
	"os"
	"path/filepath"
	"runtime"
)

// Extracts the project root directory from the current file path
func GetProjectRoot() string {
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			// Reached root without finding go.mod, return current dir as fallback
			return filepath.Dir(file)
		}
		dir = parentDir
	}
}
