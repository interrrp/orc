package main

import "github.com/BurntSushi/toml"

type Config struct {
	Services []ServiceConfig
}

type ServiceConfig struct {
	Name             string
	Command          string
	LogFile          string `toml:"log_file"`
	Mode             string
	RestartOnFailure *bool `toml:"restart_on_failure"`
}

func ReadConfig(path string) (Config, error) {
	var config Config
	_, err := toml.DecodeFile(path, &config)
	return config, err
}
