package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	configPath = "config.yaml"
)

type Config struct {

	//StoragePath   string         `yaml:"storage_path" env-required:"true"`
	//MigrationPath string         `yaml:"storage_path" env-required:"true"`
	GRPC     GRPCConfig     `yaml:"grpc"`
	Postgres PostgresConfig `yaml:"postgres"`
	Logging  Logging        `yaml:"logging"`
}

type PostgresConfig struct {
	//TODO  POSTGRES DSN
	Port     int    `yaml:"port"`
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type Logging struct {
	Filename   string `yaml:"filename"`
	MaxSize    int    `yaml:"max_size"`
	MaxAge     int    `yaml:"max_age"`
	MaxBackups int    `yaml:"max_backups"`
	LocalTime  bool   `yaml:"localtime"`
	Compress   bool   `yaml:"compress"`
}

// TODO ENV Carolos env
func New(path string) (Config, error) {
	if path == "" {
		path = configPath
	}

	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}

	defer func(file *os.File) {
		if err = file.Close(); err != nil {
			fmt.Printf("error to close config file, error: %v", err) //nolint:forbidigo
		}
	}(file)

	var config Config

	if err = yaml.NewDecoder(file).Decode(&config); err != nil {
		return Config{}, err
	}

	return config, nil
}
