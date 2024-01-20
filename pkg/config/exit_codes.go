package config

// ExitCodeBlockingError should be returned the scanner is completely
// inoperable. For example, the config is broken or it can't pull patterns
// for the first time.
const ExitCodeBlockingError = 1

// ExitCodeGeneralError should be returned when the scanner was able to run
// but there were still errors. For example it couldn't refresh the patterns
// and had to fall back on stale patterns.
const ExitCodeGeneralError = 2

// ExitCodeLeakFound should be returned when leaks are found. If there are both
// general errors and leaks, this should be returned instead of general errors.
const ExitCodeLeakFound = 3
