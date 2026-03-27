package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	// Display settings
	ShowLogo   bool   `json:"show_logo"`
	ShowColors bool   `json:"show_colors"`
	ColorTheme string `json:"color_theme"`
	Padding    int    `json:"padding"`

	// Module settings
	Modules []string `json:"modules"`
	ShowAll bool     `json:"show_all"`

	// Output settings
	OutputFormat string `json:"output_format"` // text, json, yaml
	CompactMode  bool   `json:"compact_mode"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		ShowLogo:     true,
		ShowColors:   true,
		ColorTheme:   "auto",
		Padding:      2,
		Modules:      []string{"system", "cpu", "memory", "disk"},
		ShowAll:      false,
		OutputFormat: "text",
		CompactMode:  false,
	}
}

// LoadConfig loads configuration from file
func LoadConfig(path string) (*Config, error) {
	if path == "" {
		// Use default path
		home, err := os.UserHomeDir()
		if err != nil {
			return DefaultConfig(), nil
		}
		path = filepath.Join(home, ".config", "hardfetch", "config.json")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config, path string) error {
	if path == "" {
		// Use default path
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configDir := filepath.Join(home, ".config", "hardfetch")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
		path = filepath.Join(configDir, "config.json")
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GenerateDefaultConfig generates and saves default configuration
func GenerateDefaultConfig(path string) error {
	config := DefaultConfig()
	return SaveConfig(config, path)
}
