package scanner

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	gitleaksconfig "github.com/zricethezav/gitleaks/v8/config"

	"github.com/leaktk/scanner/pkg/config"
	"github.com/leaktk/scanner/pkg/logger"
)

// Patterns acts as an abstraction for fetching different scanner patterns
// and keeping them up to date and cached
type Patterns struct {
	client         HTTPClient
	config         *config.Patterns
	gitleaksConfig *gitleaksconfig.Config
	mutex          sync.Mutex
}

// NewPatterns returns a configured instance of Patterns
func NewPatterns(cfg *config.Patterns, client HTTPClient) *Patterns {
	return &Patterns{
		client: client,
		config: cfg,
	}
}

func (p *Patterns) fetchGitleaksConfig() (string, error) {
	logger.Info("fetching gitleaks patterns")
	url, err := url.JoinPath(
		p.config.Server.URL, "patterns", "gitleaks", p.config.Gitleaks.Version,
	)

	logger.Debug("patterns url: url=%q", url)
	if err != nil {
		return "", err
	}

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	if len(p.config.Server.AuthToken) > 0 {
		logger.Debug("setting authorization header")
		request.Header.Add(
			"Authorization",
			fmt.Sprintf("Bearer %s", p.config.Server.AuthToken),
		)
	}

	response, err := p.client.Do(request)

	if err != nil {
		return "", err
	}

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(body), err
}

// gitleaksConfigModTimeExceeds returns true if the file is older than
// `modTimeLimit` seconds
func (p *Patterns) gitleaksConfigModTimeExceeds(modTimeLimit uint32) bool {
	if fileInfo, err := os.Stat(p.config.Gitleaks.ConfigPath); err == nil {
		return uint32(time.Now().Sub(fileInfo.ModTime()).Seconds()) > modTimeLimit
	}

	return true
}

// Gitleaks returns a Gitleaks config object if it's able to
func (p *Patterns) Gitleaks() (*gitleaksconfig.Config, error) {
	// Lock since this this updates the value of p.gitleaksConfig on the fly
	// and updates files on the filesystem
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.config.Autofetch && p.gitleaksConfigModTimeExceeds(p.config.RefreshAfter) {
		rawConfig, err := p.fetchGitleaksConfig()

		if err != nil {
			return p.gitleaksConfig, err
		}

		p.gitleaksConfig, err = ParseGitleaksConfig(rawConfig)
		if err != nil {
			logger.Debug("fetched config:\n%s\n", rawConfig)
			return p.gitleaksConfig, fmt.Errorf("could not parse config: error=%q", err)
		}

		if err := os.MkdirAll(filepath.Dir(p.config.Gitleaks.ConfigPath), 0700); err != nil {
			return p.gitleaksConfig, fmt.Errorf("could not create config dir: error=%q", err)
		}

		// only write the config after parsing it, that way we don't break a good
		// existing config if the server returns an invalid response
		if err := os.WriteFile(p.config.Gitleaks.ConfigPath, []byte(rawConfig), 0600); err != nil {
			return p.gitleaksConfig, fmt.Errorf("could not write config: error=%q", err)
		}
	} else if p.gitleaksConfig == nil {
		if p.gitleaksConfigModTimeExceeds(p.config.ExpiredAfter) {
			return nil, fmt.Errorf(
				"gitleaks config is expired and autofetch is disabled: config_path=%q",
				p.config.Gitleaks.ConfigPath,
			)
		}

		rawConfig, err := os.ReadFile(p.config.Gitleaks.ConfigPath)
		if err != nil {
			return p.gitleaksConfig, err
		}

		p.gitleaksConfig, err = ParseGitleaksConfig(string(rawConfig))
		if err != nil {
			logger.Debug("loaded config:\n%s\n", rawConfig)
			return p.gitleaksConfig, fmt.Errorf("could not parse config: error=%q", err)
		}
	}

	return p.gitleaksConfig, nil
}

// ParseGitleaksConfig takes a gitleaks config string and returns a config object
func ParseGitleaksConfig(rawConfig string) (*gitleaksconfig.Config, error) {
	var vc gitleaksconfig.ViperConfig

	_, err := toml.Decode(rawConfig, &vc)
	if err != nil {
		return nil, err
	}

	cfg, err := vc.Translate()

	if len(cfg.Rules) == 0 {
		return nil, fmt.Errorf("no rules found in config")
	}

	return &cfg, err
}
