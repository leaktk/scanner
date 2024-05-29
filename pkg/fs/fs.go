package fs

import (
	"os"
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
