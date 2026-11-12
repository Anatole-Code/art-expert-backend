package config_artworkserver

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	ServiceHost string
	ServicePort int
}

func FindProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "docker-compose.yml")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

func Load(root string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(filepath.Join(root, "config"))
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}
	if cfg.ServiceHost == "" || cfg.ServicePort == 0 {
		return nil, fmt.Errorf("в config.toml нужны ServiceHost и ServicePort")
	}
	return cfg, nil
}

func (cfg *Config) ListenAddress() string {
	return fmt.Sprintf("%s:%d", cfg.ServiceHost, cfg.ServicePort)
}
