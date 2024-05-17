package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/adrg/xdg"

	gitleaksconfig "github.com/leaktk/gitleaks7/v2/config"

	"github.com/leaktk/scanner/pkg/logger"
)

const nixGlobalConfigDir = "/etc/leaktk"

type (
	// Config provides a general structure to capture the config options
	// for the toolchain. This may be abstracted out to a common library in
	// the future as more components are added to the toolchain.
	Config struct {
		Logger  Logger  `toml:"logger"`
		Scanner Scanner `toml:"scanner"`
	}

	// Logger provides general logger config
	Logger struct {
		Level string `toml:"level"`
	}

	// Scanner provides scanner specific config
	Scanner struct {
		Patterns     Patterns `toml:"patterns"`
		Workdir      string   `toml:"workdir"`
		MaxScanDepth int      `toml:"max_scan_depth"`
	}

	// Patterns provides configuration for managing pattern updates
	Patterns struct {
		RefreshInterval int           `toml:"refresh_interval"`
		Server          PatternServer `toml:"server"`
		Autofetch       bool          `toml:"autofetch"`
		Gitleaks        Gitleaks      `toml:"gitleaks"`
	}

	// PatternServer provides pattern server configuration settings for the scanner
	PatternServer struct {
		URL       string `toml:"url"`
		AuthToken string `toml:"auth_token"`
	}

	// Gitleaks holds version and config information for the Gitleaks scanner
	Gitleaks struct {
		Version string `toml:"version"`
		// This is kept in memory to speed up streaming scans when using the listen
		// sub-command. Another use case would be to set up the patterns in your
		// single config.toml and turn Autofetch to false if you wanted all of the
		// config to be self-contained for use in things like a server where the
		// config file is templated out.
		Config *gitleaksconfig.Config `toml:"config"`

		ConfigPath string `toml:"config_path"`
	}
)

func leaktkConfigDir() string {
	return filepath.Join(xdg.ConfigHome, "leaktk")
}

func leaktkCacheDir() string {
	return filepath.Join(xdg.CacheHome, "leaktk")
}

func loadPatternServerAuthTokenFromFile(path string) string {
	path = filepath.Clean(path)
	logger.Debug("loading pattern-server-auth-token from %s", path)
	authTokenBytes, err := os.ReadFile(path)

	if err != nil {
		logger.Fatal("loadPatternServerAuthTokenFromFile: %s", err)
	}

	return strings.TrimSpace(string(authTokenBytes))
}

func loadPatternServerAuthToken() string {
	authTokenFromEnvVar := os.Getenv("LEAKTK_PATTERN_SERVER_AUTH_TOKEN")

	if len(authTokenFromEnvVar) > 0 {
		logger.Debug("loading pattern-server-auth-token from env var")
		return authTokenFromEnvVar
	}

	path := filepath.Join(leaktkConfigDir(), "pattern-server-auth-token")
	if _, err := os.Stat(path); err == nil {
		return loadPatternServerAuthTokenFromFile(path)
	}

	path = filepath.Join(nixGlobalConfigDir, "pattern-server-auth-token")
	if _, err := os.Stat(path); err == nil {
		return loadPatternServerAuthTokenFromFile(path)
	}

	return ""
}

func stringToBool(value string, defaultValue bool) bool {
	if len(value) == 0 {
		return defaultValue
	}

	if value == "true" || value == "1" {
		return true
	}

	return false
}

// DefaultConfig provides a fully usable instance of Config with default
// values provided
func DefaultConfig() *Config {
	return &Config{
		Logger: Logger{
			Level: "INFO",
		},
		Scanner: Scanner{
			Workdir:      filepath.Join(leaktkCacheDir(), "scanner"),
			MaxScanDepth: 0,
			Patterns: Patterns{
				Autofetch:       true,
				RefreshInterval: 60 * 60 * 12,
				Server: PatternServer{
					URL: "https://raw.githubusercontent.com/leaktk/patterns/main/target",
				},
				Gitleaks: Gitleaks{
					Version: "7.6.1",
				},
			},
		},
	}
}

// LoadConfigFromFile provides a config object with default values set plus any
// custom values pulled in from the config file
func LoadConfigFromFile(path string) (*Config, error) {
	path = filepath.Clean(path)
	logger.Debug("loading config from %s", path)
	cfg := DefaultConfig()
	_, err := toml.DecodeFile(path, cfg)

	if err != nil {
		return nil, err
	}

	// The following items take precedence over the config file

	envLoggerLevel := os.Getenv("LEAKTK_LOGGER_LEVEL")
	if len(envLoggerLevel) > 0 {
		cfg.Logger.Level = envLoggerLevel
	}

	urlFromEnvVar := os.Getenv("LEAKTK_PATTERN_SERVER_URL")
	if len(urlFromEnvVar) != 0 {
		cfg.Scanner.Patterns.Server.URL = urlFromEnvVar
	}

	// It's better to have the auth token out of the config file to make it
	// easier to write via the login command and to minimize secrets in the
	// config file. But it's still supported in the config file in case it's
	// desirable to generate one big config file for server based deployments.
	authToken := loadPatternServerAuthToken()
	if len(authToken) != 0 {
		cfg.Scanner.Patterns.Server.AuthToken = authToken
	}

	cfg.Scanner.Patterns.Autofetch = stringToBool(
		os.Getenv("LEAKTK_SCANNER_AUTOFETCH"),
		cfg.Scanner.Patterns.Autofetch,
	)

	// The folowing items are defaults built from other settings

	if len(cfg.Scanner.Patterns.Gitleaks.ConfigPath) == 0 {
		cfg.Scanner.Patterns.Gitleaks.ConfigPath = filepath.Join(
			cfg.Scanner.Workdir, "patterns", "gitleaks",
			cfg.Scanner.Patterns.Gitleaks.Version,
		)
	}

	return cfg, err
}

// LocateAndLoadConfig looks through the possible places for the config
// favoring the provided path if it is set
func LocateAndLoadConfig(path string) (*Config, error) {
	if len(path) > 0 {
		return LoadConfigFromFile(path)
	}

	if path = os.Getenv("LEAKTK_CONFIG_PATH"); len(path) > 0 {
		return LoadConfigFromFile(path)
	}

	logger.Info("loading config from %s", path)

	path = filepath.Join(leaktkConfigDir(), "config.toml")
	if _, err := os.Stat(path); err == nil {
		return LoadConfigFromFile(path)
	}

	path = filepath.Join(nixGlobalConfigDir, "config.toml")
	if _, err := os.Stat(path); err == nil {
		return LoadConfigFromFile(path)
	}

	return DefaultConfig(), nil
}
