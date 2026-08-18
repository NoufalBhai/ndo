package config

import (
	"fmt"
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"
)

// Settings holds app-level behavior — not recipes. Kept in a separate file
// (config.toml) from central.toml/.ndo.toml so recipe-parsing code never
// has to deal with non-recipe keys.
type Settings struct {
	Shell  string `toml:"shell,omitempty"`
	Editor string `toml:"editor,omitempty"`
	Color  bool   `toml:"color"`
}

type settingsFile struct {
	Settings Settings `toml:"settings"`
}

// LoadSettings reads config.toml from ndoHome. A missing file is not an
// error — it returns zero-value Settings.
func LoadSettings(ndoHome string) (Settings, error) {
	path := filepath.Join(ndoHome, "config.toml")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Settings{}, nil
	}
	if err != nil {
		return Settings{}, fmt.Errorf("reading %s: %w", path, err)
	}

	var sf settingsFile
	if err := toml.Unmarshal(data, &sf); err != nil {
		return Settings{}, fmt.Errorf("parsing %s: %w", path, err)
	}
	return sf.Settings, nil
}

// SaveSettings writes config.toml into ndoHome, creating the directory if
// needed.
func SaveSettings(ndoHome string, s Settings) error {
	if err := os.MkdirAll(ndoHome, 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", ndoHome, err)
	}
	data, err := toml.Marshal(settingsFile{Settings: s})
	if err != nil {
		return fmt.Errorf("encoding settings: %w", err)
	}
	path := filepath.Join(ndoHome, "config.toml")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}
