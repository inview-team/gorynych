package config

import (
	"fmt"
	"os"

	"github.com/inview-team/gorynych/internal/infrastructure/mongo"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Database mongo.Config `yaml:"database,omitempty"`
}

var (
	DefaultConfig Config = Config{
		Database: mongo.DefaultConfig,
	}
)

func Load(s string) (*Config, error) {
	cfg := DefaultConfig

	err := yaml.Unmarshal([]byte(s), &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func LoadFile(filename string) (*Config, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to find configuration file")
	}
	cfg, err := Load(string(content))
	if err != nil {
		return nil, fmt.Errorf("parsing YAML file %s failed: %v", filename, err)
	}
	return cfg, nil
}
