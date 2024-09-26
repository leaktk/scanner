package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileExists checks to see if a path exists and is a file
func FileExists(path string) bool {
	info, err := os.Stat(path)

	if err != nil && !os.IsNotExist(err) {
		return false
	}

	return info != nil && err == nil && !info.IsDir()
}

// PathExists checks to see if a path exists
func PathExists(path string) bool {
	info, err := os.Stat(path)

	if err != nil && !os.IsNotExist(err) {
		return false
	}

	return info != nil && err == nil
}

// CleanJoin checks to make sure that the prefix path remains after the join, this is to
// control for path traversal
func CleanJoin(prefix string, elem string) (string, error) {
	destPath := filepath.Join(prefix, elem)
	if !strings.HasPrefix(destPath, filepath.Clean(prefix)) {
		return "", fmt.Errorf("illegal file path: %s", elem)
	}
	return destPath, nil
}
