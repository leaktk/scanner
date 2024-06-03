package scanner

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

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
}

// NewPatterns returns a configured instance of Patterns
func NewPatterns(cfg *config.Patterns, client HTTPClient) *Patterns {
	return &Patterns{
		client: client,
		config: cfg,
	}
}

func (p *Patterns) fetchGitleaksConfig() (string, error) {
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

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(body), err
}

// Gitleaks returns a Gitleaks config object if it's able to
// TODO: make sure this is safe for concurrency
func (p *Patterns) Gitleaks() (*gitleaksconfig.Config, error) {
	var err error

	if p.gitleaksConfig == nil {
		var rawConfig string
		fetchedNewConfig := false

		// TODO: load patterns from FS if they exist AND are newer than the refresh time
		if contents, err := os.ReadFile(p.config.Gitleaks.ConfigPath); err == nil {
			rawConfig = string(contents)
		} else {
			if !p.config.Autofetch {
				return p.gitleaksConfig, fmt.Errorf("could not autofetch gitleaks config because autofetch is disabled")
			}

			rawConfig, err = p.fetchGitleaksConfig()
			fetchedNewConfig = true
			if err != nil {
				return p.gitleaksConfig, err
			}
		}

		p.gitleaksConfig, err = ParseGitleaksConfig(rawConfig)
		if err != nil {
			logger.Debug("returned gitleaks config:\n%v", rawConfig)
			return p.gitleaksConfig, err
		}

		err = os.MkdirAll(filepath.Dir(p.config.Gitleaks.ConfigPath), 0700)
		if err != nil {
			return p.gitleaksConfig, err
		}

		if fetchedNewConfig {
			configFile, err := os.Create(p.config.Gitleaks.ConfigPath)
			if err != nil {
				return p.gitleaksConfig, err
			}
			defer configFile.Close()

			_, err = configFile.WriteString(rawConfig)
			if err != nil {
				return p.gitleaksConfig, err
			}
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
	return &cfg, err
}
