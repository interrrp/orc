package main

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type config struct {
	Services []service
}

func readConfig(path string) (config, error) {
	var cfg config
	_, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("failed to read config from %s: %w", path, err)
	}
	return cfg, nil
}
