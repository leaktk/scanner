package config

import (
	"errors"
	"github.com/BurntSushi/toml"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
)

type (
	// Config provides a general structure to capture the config options
	// for the toolchain. This may be abstracted out to a common library in
	// the future as more components are added to the toolchain.
	Config struct {
		Logger  Logger  `toml:"logger"`
		Scanner Scanner `toml:"scanner"`
	}

	// Logger provides general logging config
	Logger struct {
		Level string `toml:"level"`
	}

	// Scanner provides scanner specific config
	Scanner struct {
		Workdir  string   `toml:"workdir"`
		Gitleaks Gitleaks `toml:"gitleaks"`
		Patterns Patterns `toml:"patterns"`
	}

	// Gitleaks configures the gitleaks subscanner
	Gitleaks struct {
		Version string `toml:"version"`
	}

	// Patterns provides configuration for managing pattern updates
	Patterns struct {
		RefreshInterval int           `toml:"refresh_interval"`
		Server          PatternServer `toml:"server"`
	}

	// PatternServer provides pattern server configuration settings for the scanner
	PatternServer struct {
		URL       string `toml:"url"`
		AuthToken string `toml:"auth_token"`
	}
)

func leaktkConfigDir() string {
	return filepath.Join(xdg.ConfigHome, "leaktk")
}

func leaktkCacheDir() string {
	return filepath.Join(xdg.CacheHome, "leaktk")
}

func defaultScannerWorkdir() string {
	return filepath.Join(leaktkCacheDir(), "scanner")
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

func validateLoggingLevel(config *Config) error {
	switch level := config.Logger.Level; level {
	case "ERROR":
		return nil
	case "WARN":
		return nil
	case "INFO":
		return nil
	case "DEBUG":
		return nil
	case "TRACE":
		return nil
	default:
		return errors.New(level + " is an invalid log level")
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
			Workdir: defaultScannerWorkdir(),
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
	config := DefaultConfig()
	_, err := toml.DecodeFile(filepath.Clean(path), config)

	if err != nil {
		return nil, err
	}

	err = validateLoggingLevel(config)

	if err != nil {
		return nil, err
	}

	return config, err
}
