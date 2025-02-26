package config

// ExitCodeBlockingError should be returned the scanner is completely
// inoperable. For example, the config is broken or it can't pull patterns
// for the first time.
const ExitCodeBlockingError = 1

// LeakExitCode is returned when a scan returns results
const LeakExitCode = 42
