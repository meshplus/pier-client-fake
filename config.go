package main

import (
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const (
	configName = "fake.toml"
)

type Config struct {
	Fake Fake `toml:"fake" json:"fake"`
}

type Fake struct {
	Name string `toml:"name" json:"name"`
	To   string `toml:"to" json:"to"`
	Tps  int64  `toml:"tps" json:"tps"`
}

func defaultConfig() *Config {
	return &Config{
		Fake: Fake{
			Name: "Fake",
		},
	}
}

func UnmarshalConfig(configRoot string) (*Config, error) {
	viper.SetConfigFile(filepath.Join(configRoot, configName))
	viper.SetConfigType("toml")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("FAKE")
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	config := defaultConfig()

	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	return config, nil
}
