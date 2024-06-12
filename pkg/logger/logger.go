package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// LogLevel is used to determine which log severities should actually log
type LogLevel int

// LogFormat is used to set the how the log messages should be displayed
type LogFormat int

const (
	// NOTSET will log everything
	NOTSET LogLevel = 0
	// DEBUG will enable these logs and higer
	DEBUG LogLevel = 10
	// INFO will enable these logs and higer
	INFO LogLevel = 20
	// WARNING will enable these logs and higer
	WARNING LogLevel = 30
	// ERROR will enable these logs and higer
	ERROR LogLevel = 40
	// CRITICAL will enable these logs and higer
	CRITICAL LogLevel = 50
)

const (
	// JSON displays the logs as JSON dicts
	JSON LogFormat = 0
	// HUMAN displays the logs in a way that's nice for humans to read
	HUMAN LogFormat = 1
)

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
func Debug(msg string, a ...any) {
	if currentLogLevel > DEBUG {
		return
	}

	log.Println(Entry{
		Time:     time.Now().UTC().Format(time.RFC3339),
		Severity: "DEBUG",
		Message:  fmt.Sprintf(msg, a...),
	})
}

// Info emits an INFO level log
func Info(msg string, a ...any) {
	if currentLogLevel > INFO {
		return
	}

	log.Println(Entry{
		Time:     time.Now().UTC().Format(time.RFC3339),
		Severity: "INFO",
		Message:  fmt.Sprintf(msg, a...),
	})
}

// Warning emits an WARNING level log
func Warning(msg string, a ...any) {
	if currentLogLevel > WARNING {
		return
	}

	log.Println(Entry{
		Time:     time.Now().UTC().Format(time.RFC3339),
		Severity: "WARNING",
		Message:  fmt.Sprintf(msg, a...),
	})
}

// Error emits an ERROR level log
func Error(msg string, a ...any) {
	if currentLogLevel > ERROR {
		return
	}

	log.Println(Entry{
		Time:     time.Now().UTC().Format(time.RFC3339),
		Severity: "ERROR",
		Message:  fmt.Errorf(msg, a...).Error(),
	})
}

// Fatal emits an CRITICAL level log and stops the program
func Fatal(msg string, a ...any) {
	log.Fatal(Entry{
		Time:     time.Now().UTC().Format(time.RFC3339),
		Severity: "CRITICAL",
		Message:  fmt.Errorf(msg, a...).Error(),
	})
}
