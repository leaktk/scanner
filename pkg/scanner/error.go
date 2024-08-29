package scanner

import "fmt"

type LeakTKError struct {
	Fatal   bool      `json:"fatal"`
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func (e LeakTKError) Error() string {
	return e.String()
}

func (e LeakTKError) String() string {
	fatal := ""
	if e.Fatal {
		fatal = "fatal "
	}
	return fmt.Sprintf("%serror occured, code %d (%s): %s", fatal, e.Code, errorNames[e.Code], e.Message)
}

type ErrorCode int

const (
	NoError = iota
	CloneError
	ScanError
	ResourceCleanupError
)

var errorNames = [...]string{"NoError", "CloneError", "ScanError", "ResourceCleanupError"}
