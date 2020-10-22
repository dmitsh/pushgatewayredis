package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"

	"github.com/dmitsh/pushgatewayredis/pkg/redis"
)

type Config struct {
	Port          int    `yaml:"port"`
	MetricsPath   string `yaml:"metrics_path"`
	TelemetryPath string `yaml:"telemetry_path"`
	IngestPath    string `yaml:"ingest_path"`
	TLSEnabled    bool   `yaml:"tls_enabled"`
	TLSKeyPath    string `yaml:"tls_key_path"`
	TLSCertPath   string `yaml:"tls_cert_path"`

	RedisConfig redis.RedisConfig `yaml:"redis"`
}

// LoadFile parses the given YAML file into a Config.
func (cfg *Config) LoadFile(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	if err = yaml.Unmarshal(data, cfg); err != nil {
		return err
	}
	return nil
}
