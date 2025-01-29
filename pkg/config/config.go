package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/adrg/xdg"

	"github.com/leaktk/scanner/pkg/fs"
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
		CloneTimeout        uint16   `toml:"clone_timeout"`
		CloneWorkers        uint16   `toml:"clone_workers"`
		IncludeResponseLogs bool     `toml:"include_response_logs"`
		MaxDecodeDepth      uint16   `toml:"max_decode_depth"`
		MaxScanDepth        uint16   `toml:"max_scan_depth"`
		Patterns            Patterns `toml:"patterns"`
		ScanWorkers         uint16   `toml:"scan_workers"`
		Workdir             string   `toml:"workdir"`
	}

	// Patterns provides configuration for managing pattern updates
	Patterns struct {
		Autofetch    bool          `toml:"autofetch"`
		ExpiredAfter uint32        `toml:"expired_after"`
		Gitleaks     Gitleaks      `toml:"gitleaks"`
		RefreshAfter uint32        `toml:"refresh_after"`
		Server       PatternServer `toml:"server"`
	}

	// Gitleaks holds version and config information for the Gitleaks scanner
	Gitleaks struct {
		Version    string `toml:"version"`
		ConfigPath string `toml:"config_path"`
	}

	// PatternServer provides pattern server configuration settings for the scanner
	PatternServer struct {
		AuthToken string `toml:"auth_token"`
		URL       string `toml:"url"`
	}
)

// Make sure that any config returned to the code goes through this function
func setMissingValues(cfg *Config) *Config {
	envLoggerLevel := os.Getenv("LEAKTK_LOGGER_LEVEL")
	if len(envLoggerLevel) > 0 {
		cfg.Logger.Level = envLoggerLevel
	}

	urlFromEnvVar := os.Getenv("LEAKTK_PATTERN_SERVER_URL")
	if len(urlFromEnvVar) != 0 {
		cfg.Scanner.Patterns.Server.URL = urlFromEnvVar
	}

	cfg.Scanner.Patterns.Autofetch = stringToBool(
		os.Getenv("LEAKTK_SCANNER_AUTOFETCH"),
		cfg.Scanner.Patterns.Autofetch,
	)

	// It's better to have the auth token out of the config file to make it
	// easier to write via the login command and to minimize secrets in the
	// config file. But it's still supported in the config file in case it's
	// desirable to generate one big config file for server based deployments.
	if authToken := loadPatternServerAuthToken(); len(authToken) != 0 {
		cfg.Scanner.Patterns.Server.AuthToken = authToken
	}

	if len(cfg.Scanner.Patterns.Gitleaks.ConfigPath) == 0 {
		cfg.Scanner.Patterns.Gitleaks.ConfigPath = filepath.Join(
			cfg.Scanner.Workdir, "patterns", "gitleaks",
			cfg.Scanner.Patterns.Gitleaks.Version,
		)
	}

	return cfg
}

func leaktkConfigDir() string {
	path := filepath.Join(xdg.ConfigHome, "leaktk")

	if err := os.MkdirAll(path, 0770); err != nil {
		logger.Error("could not create dir: path=%q", path)
	}

	return path
}

func leaktkCacheDir() string {
	path := filepath.Join(xdg.CacheHome, "leaktk")

	if err := os.MkdirAll(path, 0770); err != nil {
		logger.Error("could not create dir: path=%q", path)
	}

	return path
}

func loadPatternServerAuthTokenFromFile(path string) string {
	path = filepath.Clean(path)
	logger.Debug("loading pattern-server-auth-token: path=%q", path)
	authTokenBytes, err := os.ReadFile(path)

	if err != nil {
		logger.Fatal("could not load pattern server auth token: error=%q", err)
	}

	return strings.TrimSpace(string(authTokenBytes))
}

func localPatternServerAuthTokenPath() string {
	return filepath.Join(leaktkConfigDir(), "pattern-server-auth-token")
}

func loadPatternServerAuthToken() string {
	authTokenFromEnvVar := os.Getenv("LEAKTK_PATTERN_SERVER_AUTH_TOKEN")

	if len(authTokenFromEnvVar) > 0 {
		logger.Debug("loading pattern-server-auth-token from env var")
		return authTokenFromEnvVar
	}

	path := localPatternServerAuthTokenPath()
	if fs.FileExists(path) {
		return loadPatternServerAuthTokenFromFile(path)
	}

	path = filepath.Join(nixGlobalConfigDir, "pattern-server-auth-token")
	if fs.FileExists(path) {
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
			CloneTimeout:        0,
			CloneWorkers:        1,
			IncludeResponseLogs: false,
			MaxScanDepth:        0,
			ScanWorkers:         1,
			Workdir:             filepath.Join(leaktkCacheDir(), "scanner"),
			MaxDecodeDepth:      8,
			Patterns: Patterns{
				Autofetch:    true,
				ExpiredAfter: 60 * 60 * 12 * 14, // 7 days
				RefreshAfter: 60 * 60 * 12,      // 12 hours
				Gitleaks: Gitleaks{
					Version: "8.18.2",
				},
				Server: PatternServer{
					URL: "https://raw.githubusercontent.com/leaktk/patterns/main/target",
				},
			},
		},
	}
}

// LoadConfigFromFile provides a config object with default values set plus any
// custom values pulled in from the config file
func LoadConfigFromFile(path string) (*Config, error) {
	path = filepath.Clean(path)
	logger.Debug("loading config: path=%q", path)
	cfg := DefaultConfig()
	_, err := toml.DecodeFile(path, cfg)

	if err != nil {
		return nil, err
	}

	return setMissingValues(cfg), err
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

	if len(path) > 0 {
		logger.Debug("loading config: path=%q", path)
	} else {
		logger.Debug("using default config")
	}

	path = filepath.Join(leaktkConfigDir(), "config.toml")
	if fs.FileExists(path) {
		return LoadConfigFromFile(path)
	}

	path = filepath.Join(nixGlobalConfigDir, "config.toml")
	if fs.FileExists(path) {
		return LoadConfigFromFile(path)
	}

	return setMissingValues(DefaultConfig()), nil
}

// SavePatternServerAuthToken saves the token in the path where it should go
func SavePatternServerAuthToken(authToken string) error {
	path := localPatternServerAuthTokenPath()

	if err := os.WriteFile(path, []byte(strings.TrimSpace(authToken)), 0660); err != nil {
		return err
	}

	return nil
}

// RemovePatternServerAuthToken deletes the auth token
func RemovePatternServerAuthToken() error {
	path := localPatternServerAuthTokenPath()

	if fs.FileExists(path) {
		if err := os.Remove(path); err != nil {
			return err
		}
	}

	return nil
}
