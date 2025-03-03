package response

import "fmt"

// LeakTKError expands a normal error to provide additional meta data
type LeakTKError struct {
	Fatal   bool      `json:"fatal"`
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

// Error is defined to implement the error interface
func (e LeakTKError) Error() string {
	return e.String()
}

// String provides a string representation of the error
func (e LeakTKError) String() string {
	fatal := ""

	if e.Fatal {
		fatal = "fatal "
	}

	return fmt.Sprintf("%serror occurred, code %d (%s): %s", fatal, e.Code, errorNames[e.Code], e.Message)
}

// ErrorCode defines the set of error codes that can be set on a LeakTKError
type ErrorCode int

const (
	// NoErrorCode means the error code hasn't been set
	NoErrorCode = iota
	// CloneError means we were unable to successfully clone the resource
	CloneError
	// ScanError means there was some issue scanning the cloned resource
	ScanError
	// ResourceCleanupError means we couldn't remove the resources that were cloned after a scan
	ResourceCleanupError
	// LocalScanDisabled means local scans are not enabled
	LocalScanDisabled
)

var errorNames = [...]string{"NoErrorCode", "CloneError", "ScanError", "ResourceCleanupError"}
