package scanner

import (
	"crypto/sha256"
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
	client             HTTPClient
	config             *config.Patterns
	gitleaksConfigHash [32]byte
	gitleaksConfig     *gitleaksconfig.Config
	mutex              sync.Mutex
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
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: status_code=%d", response.StatusCode)
	}

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
		return uint32(time.Since(fileInfo.ModTime()).Seconds()) > modTimeLimit
	}

	return true
}

// Gitleaks returns a Gitleaks config object if it's able to
func (p *Patterns) Gitleaks() (*gitleaksconfig.Config, error) {
	// Lock since this updates the value of p.gitleaksConfig on the fly
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
			logger.Debug("fetched config:\n%s", rawConfig)
			return p.gitleaksConfig, fmt.Errorf("could not parse config: error=%q", err)
		}

		if err := os.MkdirAll(filepath.Dir(p.config.Gitleaks.ConfigPath), 0700); err != nil {
			return p.gitleaksConfig, fmt.Errorf("could not create config dir: error=%q", err)
		}

		// only write the config after parsing it, that way we don't break a good
		// existing config if the server returns an invalid response
		if err := os.WriteFile(p.config.Gitleaks.ConfigPath, []byte(rawConfig), 0600); err != nil {
			return p.gitleaksConfig, fmt.Errorf("could not write config: path=%q error=%q", p.config.Gitleaks.ConfigPath, err)
		}
		p.updateGitleaksConfigHash(sha256.Sum256([]byte(rawConfig)))
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
		p.updateGitleaksConfigHash(sha256.Sum256(rawConfig))
	}

	return p.gitleaksConfig, nil
}

// GitleaksConfigHash returns the sha256 hash for the current gitleaks config
func (p *Patterns) GitleaksConfigHash() string {
	return fmt.Sprintf("%x", p.gitleaksConfigHash)
}

// updateGitleaksConfigHash updated value and logs only on a change
func (p *Patterns) updateGitleaksConfigHash(hash [32]byte) {
	if hash != p.gitleaksConfigHash {
		p.gitleaksConfigHash = hash
		logger.Info("updated gitleaks patterns: hash=%s", p.GitleaksConfigHash())
	}
}

func invalidConfig(cfg *gitleaksconfig.Config) bool {
	al := &cfg.Allowlist

	// Make sure something the scanner can use is set
	return len(cfg.Rules) == 0 &&
		len(al.Commits) == 0 &&
		len(al.Description) == 0 &&
		len(al.Paths) == 0 &&
		len(al.Regexes) == 0 &&
		len(al.StopWords) == 0
}

// ParseGitleaksConfig takes a gitleaks config string and returns a config object
func ParseGitleaksConfig(rawConfig string) (glc *gitleaksconfig.Config, err error) {
	var vc gitleaksconfig.ViperConfig

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("gitleaks config is invalid: error=%q", r)
		}
	}()

	_, err = toml.Decode(rawConfig, &vc)
	if err != nil {
		return nil, err
	}

	cfg, err := vc.Translate()

	if invalidConfig(&cfg) {
		return nil, fmt.Errorf("invalid config")
	}

	return &cfg, err
}
