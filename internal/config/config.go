package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	DefaultModel string `json:"default_model"`
	OllamaURL    string `json:"ollama_url"`
	Debug        bool   `json:"debug"`
}

func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".sidekick", "config.json"), nil
}

func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return GetDefault(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return GetDefault(), nil
		}
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) Save() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func GetDefault() *Config {
	return &Config{
		DefaultModel: "qwen2.5-coder:14b",
		OllamaURL:    "http://localhost:11434",
		Debug:        false,
	}
}

func (c *Config) Display() {
	fmt.Printf("  Model: %s | URL: %s | Debug: %v\n",
		c.DefaultModel, c.OllamaURL, c.Debug)
}
