package config

import (
	"github.com/BurntSushi/toml"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type (
	// Config provides a general structure to capture the config options
	// for the toolchain. This may be abstracted out to a common library in
	// the future as more components are added to the toolchain.
	Config struct {
		Logger  Logger
		Scanner Scanner
	}

	// Logger provides general logging config
	Logger struct {
		Level string
	}

	// Scanner provides scanner specific config
	Scanner struct {
		Workdir  string
		Gitleaks Gitleaks
		Patterns Patterns
	}

	// Gitleaks configures the gitleaks subscanner
	Gitleaks struct {
		Version string
	}

	// Patterns provides configuration for managing pattern updates
	Patterns struct {
		RefreshInterval int
		Server          PatternServer
	}

	// PatternServer provides pattern server configuration settings for the scanner
	PatternServer struct {
		URL string
		// AuthToken is optional and is only needed if the server requires one
		AuthToken string
	}
)

func xdgDir(defaultName, envVar string) string {
	xdgDirFromEnvVar := os.Getenv(envVar)

	if len(xdgDirFromEnvVar) > 0 {
		return filepath.Clean(xdgDirFromEnvVar)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("from xdgDir: %v", err)
	}

	return filepath.Join(homeDir, defaultName)
}

func xdgConfigHome() string {
	return xdgDir(".config", "XDG_CONFIG_HOME")
}

func xdgCacheHome() string {
	return xdgDir(".cache", "XDG_CACHE_HOME")
}

func leaktkConfigDir() string {
	return filepath.Join(xdgConfigHome(), "leaktk")
}

func defaultPatternServerAuthToken() string {
	authTokenFromEnvVar := os.Getenv("LEAKTK_PATTERN_SERVER_AUTH_TOKEN")

	if len(authTokenFromEnvVar) > 0 {
		return authTokenFromEnvVar
	}

	authTokenFilePath := filepath.Join(leaktkConfigDir(), "pattern-server-auth-token")

	if _, err := os.Stat(authTokenFilePath); err == nil {
		authTokenBytes, err := os.ReadFile(authTokenFilePath)

		if err != nil {
			log.Fatalf("from defaultPatternServerAuthToken: %v", err)
		}

		return strings.TrimSpace(string(authTokenBytes))
	}

	return ""
}

func defaultPatternServerURL() string {
	urlFromEnvVar := os.Getenv("LEAKTK_PATTERN_SERVER_URL")

	if len(urlFromEnvVar) > 0 {
		return urlFromEnvVar
	}

	return "https://raw.githubusercontent.com/leaktk/patterns/main/target"
}

func validateLoggingLevel(config *Config) {
	switch level := config.Logger.Level; level {
	case "ERROR":
	case "WARN":
	case "INFO":
	case "DEBUG":
	case "TRACE":
	default:
		log.Fatalf("%s is an invalid log level", level)
	}
}

// DefaultConfig provides a fully usable instance of Config with default
// values provided
func DefaultConfig() *Config {
	return &Config{
		Logger: Logger{
			Level: "INFO",
		},
		Scanner: Scanner{
			Gitleaks: Gitleaks{
				Version: "7.6.1",
			},
			Patterns: Patterns{
				RefreshInterval: 60 * 60 * 12,
				Server: PatternServer{
					URL:       defaultPatternServerURL(),
					AuthToken: defaultPatternServerAuthToken(),
				},
			},
		},
	}
}

// LoadConfigFromFile provides a config object with default values set plus any
// custom values pulled in from the config file
func LoadConfigFromFile(path string) (*Config, error) {
	path = filepath.Clean(path)
	config := DefaultConfig()
	_, err := toml.DecodeFile(path, config)

	if err != nil {
		return nil, err
	}

	validateLoggingLevel(config)

	return config, err
}
