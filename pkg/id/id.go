package id

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// ID returns an unique ID based on the parts passed in. If none are passed in
// then ID will be random. The ID will be a fixed length hexadecimal characters.
func ID(parts ...string) string {
	var data []byte

	if len(parts) > 0 {
		data = []byte(strings.Join(parts, "\n"))
	} else {
		data = []byte(uuid.New().String())
	}

	return fmt.Sprintf("%x", sha256.Sum256(data))
}
