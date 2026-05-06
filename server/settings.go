package server

import (
	"os"
	"path/filepath"
)

// Settings holds user-configurable options persisted to ~/.skills-rpg.conf.
type Settings struct {
	Language string `yaml:"language"` // "en" (default) | "ja"
}

func (s *Settings) setDefaults() {
	if s.Language == "" {
		s.Language = "en"
	}
}

func defaultSettingsPath() string {
	return filepath.Join(os.Getenv("HOME"), ".skills-rpg.conf")
}

func loadSettings(path string) (Settings, error) {
	var cfg Settings
	err := readYAML(path, &cfg)
	if err != nil {
		if os.IsNotExist(err) {
			cfg.setDefaults()
			return cfg, nil
		}
		return cfg, err
	}
	cfg.setDefaults()
	return cfg, nil
}

func saveSettings(path string, cfg Settings) error {
	return writeYAMLAtomic(path, cfg)
}
