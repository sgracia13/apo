// Package config handles application configuration.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	DefaultAPIURL     = "https://dev.azure.com"
	DefaultAPIVersion = "7.1"
	DefaultTimeout    = 30 * time.Second
)

// Config holds the application configuration.
type Config struct {
	Organization string `json:"organization"`
	Project      string `json:"project"`
	PAT          string `json:"pat"`
	APIURL       string `json:"api_url,omitempty"`
	APIVersion   string `json:"api_version,omitempty"`
}

// Load reads configuration from file and environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		APIURL:     DefaultAPIURL,
		APIVersion: DefaultAPIVersion,
	}

	configPath := GetConfigPath()
	if data, err := os.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing config file: %w", err)
		}
	}

	// Environment overrides
	if org := os.Getenv("AZURE_DEVOPS_ORG"); org != "" {
		cfg.Organization = org
	}
	if project := os.Getenv("AZURE_DEVOPS_PROJECT"); project != "" {
		cfg.Project = project
	}
	if pat := os.Getenv("AZURE_DEVOPS_PAT"); pat != "" {
		cfg.PAT = pat
	}
	if url := os.Getenv("AZURE_DEVOPS_URL"); url != "" {
		cfg.APIURL = url
	}
	if version := os.Getenv("AZURE_DEVOPS_API_VERSION"); version != "" {
		cfg.APIVersion = version
	}

	if cfg.APIURL == "" {
		cfg.APIURL = DefaultAPIURL
	}
	if cfg.APIVersion == "" {
		cfg.APIVersion = DefaultAPIVersion
	}

	return cfg, nil
}

// Save writes the configuration to the config file.
func (c *Config) Save() error {
	configPath := GetConfigPath()
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	return os.WriteFile(configPath, data, 0600)
}

// Validate checks that required configuration is present.
func (c *Config) Validate() error {
	if c.Organization == "" {
		return fmt.Errorf("organization is required")
	}
	if c.PAT == "" {
		return fmt.Errorf("personal access token (PAT) is required")
	}
	return nil
}

// ValidateWithProject validates including project requirement.
func (c *Config) ValidateWithProject() error {
	if err := c.Validate(); err != nil {
		return err
	}
	if c.Project == "" {
		return fmt.Errorf("project is required")
	}
	return nil
}

// GetConfigPath returns the path to the config file.
func GetConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".config", "apo", "config.json")
}
