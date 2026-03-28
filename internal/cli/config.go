package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// GenerateDefaultConfig generates and saves default configuration
func GenerateDefaultConfig(path string) error {
	config := map[string]interface{}{
		"show_colors":   true,
		"color_theme":   "auto",
		"padding":       2,
		"modules":       []string{"system", "cpu", "memory", "disk"},
		"show_all":      false,
		"output_format": "text",
		"compact_mode":  false,
	}

	if path == "" {
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
