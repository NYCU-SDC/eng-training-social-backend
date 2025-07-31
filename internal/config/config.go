package config

import (
	"errors"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"os"
)

const DefaultSecret = "default-secret"

var ErrDatabaseURLRequired = errors.New("database_url is required")

type Config struct {
	Debug              bool   `yaml:"debug"              envconfig:"DEBUG"`
	Host               string `yaml:"host"               envconfig:"HOST"`
	Port               string `yaml:"port"               envconfig:"PORT"`
	BaseURL            string `yaml:"base_url"           envconfig:"BASE_URL"`
	Secret             string `yaml:"secret"             envconfig:"SECRET"`
	DatabaseURL        string `yaml:"database_url"       envconfig:"DATABASE_URL"`
	MigrationSource    string `yaml:"migration_source"   envconfig:"MIGRATION_SOURCE"`
	GoogleClientID     string `yaml:"google_client_id"   envconfig:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `yaml:"google_client_secret" envconfig:"GOOGLE_CLIENT_SECRET"`
}

func (c Config) Validate() error {
	if c.DatabaseURL == "" {
		return ErrDatabaseURLRequired
	}

	return nil
}

type LogBuffer struct {
	buffer []logEntry
}

type logEntry struct {
	msg  string
	err  error
	meta map[string]string
}

func NewConfigLogger() *LogBuffer {
	return &LogBuffer{}
}

func (cl *LogBuffer) Warn(msg string, err error, meta map[string]string) {
	cl.buffer = append(cl.buffer, logEntry{msg: msg, err: err, meta: meta})
}

func (cl *LogBuffer) FlushToZap(logger *zap.Logger) {
	for _, e := range cl.buffer {
		var fields []zap.Field
		if e.err != nil {
			fields = append(fields, zap.Error(e.err))
		}
		for k, v := range e.meta {
			fields = append(fields, zap.String(k, v))
		}
		logger.Warn(e.msg, fields...)
	}
	cl.buffer = nil
}

func Load() (Config, *LogBuffer) {
	logger := NewConfigLogger()

	config := &Config{
		Debug:           false,
		Host:            "localhost",
		Port:            "8080",
		Secret:          DefaultSecret,
		DatabaseURL:     "",
		MigrationSource: "file://internal/database/migrations",
	}

	var err error

	config, err = FromFile("config.yaml", config)
	if err != nil {
		logger.Warn("Failed to load config from file", err, map[string]string{"path": "config.yaml"})
	}

	return *config, logger
}

func FromFile(filePath string, config *Config) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return config, err
	}
	defer file.Close()

	fileConfig := Config{}
	if err := yaml.NewDecoder(file).Decode(&fileConfig); err != nil {
		return config, err
	}

	return &fileConfig, nil
}
