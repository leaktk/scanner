package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/rs/zerolog"
)

func init() {
	// Disable logging for gitleaks
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

// LogLevel is used to determine which log severities should actually log
type LogLevel int

// LogFormat is used to set the how the log messages should be displayed
type LogFormat int

const (
	// NOTSET will log everything
	NOTSET LogLevel = 0
	// DEBUG will enable these logs and higher
	DEBUG LogLevel = 10
	// INFO will enable these logs and higher
	INFO LogLevel = 20
	// WARNING will enable these logs and higher
	WARNING LogLevel = 30
	// ERROR will enable these logs and higher
	ERROR LogLevel = 40
	// CRITICAL will enable these logs and higher
	CRITICAL LogLevel = 50
)

const (
	// JSON displays the logs as JSON dicts
	JSON LogFormat = 0
	// HUMAN displays the logs in a way that's nice for humans to read
	HUMAN LogFormat = 1
)

// LogCode defines the set of  codes that can be set on an Entry
type LogCode int

const (
	// NoCode means the entry code hasn't been set
	NoCode = iota
	// CloneError means we were unable to successfully clone the resource
	CloneError
	// ScanError means there was some issue scanning the cloned resource
	ScanError
	// ResourceCleanupError means we couldn't remove the resources that were cloned after a scan
	ResourceCleanupError
	// CommandError means there was an error with an external command
	CommandError
	// CloneDetail are log entries that are informational
	CloneDetail
	// ScanDetail are log entries that are informational
	ScanDetail
)

var logCodeNames = [...]string{"NoCode", "CloneError", "ScanError", "ResourceCleanupError", "CommandError", "CloneDetail", "ScanDetail"}

func (code LogCode) String() string {
	return logCodeNames[code]
}

// String renders a LogLevel as its string value
func (l LogLevel) String() string {
	switch l {
	case NOTSET:
		return "NOTSET"
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "WARNING"
	case CRITICAL:
		return "CRITICAL"
	default:
		return "INVALID"
	}
}

var currentLogLevel = INFO
var currentLogFormat = HUMAN

// Entry defines a log entry
type Entry struct {
	Time     string `json:"time"`
	Severity string `json:"severity"`
	Code     string `json:"code,omitempty"`
	Message  string `json:"message"`
}

// String renders a log entry structure to the JSON format
func (e Entry) String() string {
	if e.Severity == "" {
		e.Severity = "INFO"
	}

	switch currentLogFormat {
	case HUMAN:
		return fmt.Sprintf("[%s] %s", e.Severity, e.Message)

	case JSON:
		out, err := json.Marshal(e)

		if err != nil {
			log.Printf("json.Marshal: %v", err)
		}

		return string(out)

	default:
		return e.Message
	}

}

func init() {
	// Disable log prefixes such as the default timestamp.
	// Prefix text prevents the message from being parsed as JSON.
	// A timestamp is added when shipping logs to Cloud Logging.
	log.SetFlags(0)
}

// SetLoggerFormat adjusts the format Entry uses when calling String() on it
func SetLoggerFormat(logFormat LogFormat) error {
	switch logFormat {
	case JSON:
		currentLogFormat = JSON
	case HUMAN:
		currentLogFormat = HUMAN
	default:
		return fmt.Errorf("invalid log format: log_format=%v", logFormat)
	}

	return nil
}

// SetLoggerLevel takes the string version of the name and sets the current level
func SetLoggerLevel(levelName string) error {
	switch levelName {
	case "DEBUG":
		currentLogLevel = DEBUG
	case "INFO":
		currentLogLevel = INFO
	case "WARNING":
		currentLogLevel = WARNING
	case "ERROR":
		currentLogLevel = ERROR
	case "CRITICAL":
		currentLogLevel = CRITICAL
	default:
		return fmt.Errorf("invalid log level: level=%q", levelName)
	}

	return nil
}

// GetLoggerLevel returns the current logger level
func GetLoggerLevel() LogLevel {
	return currentLogLevel
}

// Debug emits an DEBUG level log
func Debug(msg string, a ...any) *Entry {
	if currentLogLevel > DEBUG {
		return nil
	}
	entry := Entry{
		Time:     time.Now().UTC().Format(time.RFC3339),
		Severity: "DEBUG",
		Message:  fmt.Sprintf(msg, a...),
	}
	log.Println(entry)
	return &entry
}

// Info emits an INFO level log
func Info(msg string, a ...any) *Entry {
	if currentLogLevel > INFO {
		return nil
	}
	entry := Entry{
		Time:     time.Now().UTC().Format(time.RFC3339),
		Severity: "INFO",
		Message:  fmt.Sprintf(msg, a...),
	}
	log.Println(entry)
	return &entry
}

// Warning emits an WARNING level log
func Warning(msg string, a ...any) *Entry {
	if currentLogLevel > WARNING {
		return nil
	}
	entry := Entry{
		Time:     time.Now().UTC().Format(time.RFC3339),
		Severity: "WARNING",
		Message:  fmt.Sprintf(msg, a...),
	}
	log.Println(entry)
	return &entry
}

// Error emits an ERROR level log
func Error(msg string, a ...any) *Entry {
	if currentLogLevel > ERROR {
		return nil
	}
	entry := Entry{
		Time:     time.Now().UTC().Format(time.RFC3339),
		Severity: "ERROR",
		Message:  fmt.Errorf(msg, a...).Error(),
	}
	log.Println(entry)
	return &entry
}

// Critical emits an CRITICAL level log
func Critical(msg string, a ...any) *Entry {
	if currentLogLevel > CRITICAL {
		return nil
	}
	entry := Entry{
		Time:     time.Now().UTC().Format(time.RFC3339),
		Severity: "CRITICAL",
		Message:  fmt.Errorf(msg, a...).Error(),
	}
	log.Println(entry)
	return &entry
}

// Fatal emits an CRITICAL level log and stops the program
func Fatal(msg string, a ...any) {
	log.Fatal(Entry{
		Time:     time.Now().UTC().Format(time.RFC3339),
		Severity: "CRITICAL",
		Message:  fmt.Errorf(msg, a...).Error(),
	})
}
