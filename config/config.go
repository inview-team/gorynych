package config

import (
	"fmt"
	"os"

	"github.com/inview-team/gorynych/internal/infrastructure/mongo"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Providers []ObjectStorage `yaml:"providers,omitempty"`
	Storage   mongo.Config    `yaml:"storage,omitempty"`
}

var (
	DefaultConfig Config = Config{
		Storage: mongo.DefaultConfig,
	}
)

type ObjectStorage struct {
	Type         string `yaml:"type"`
	AccessKeyID  string `yaml:"access_key_id"`
	AccessSecret string `yaml:"access_secret"`
}

type Provider string

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
