package configs

import (
	"fmt"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

type Config struct {
	Server  ServerConfig   `koanf:"server"`
	DB      DatabaseConfig `koanf:"postgres"`
	Redis   RedisConfig    `koanf:"redis"`
	Broker  BrokerConfig   `koanf:"broker"`
	HotSpot HotSpotConfig  `koanf:"hotspot"`
}

type ServerConfig struct {
	Port string `koanf:"port"`
}

type Log struct {
	Level string `koanf:"level"`
}

type DatabaseConfig struct {
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	User     string `koanf:"user"`
	Password string `koanf:"password"`
	DB       string `koanf:"db"`
	SSLMode  string `koanf:"sslmode"`
}

type RedisConfig struct {
	Addr     string `koanf:"addr"`
	Password string `koanf:"password"`
	DB       int    `koanf:"db"`
}

type BrokerConfig struct {
	URL        string `koanf:"url"`
	Exchange   string `koanf:"exchange"`
	IntraQueue string `koanf:"intra_queue"`
	InterQueue string `koanf:"inter_queue"`
}

type HotSpotConfig struct {
	Resolution       int `koanf:"resolution"`
	DecayHalfLifeMin int `koanf:"decay_half_life_min"`
	Threshold        int `koanf:"threshold"`
}

func LoadConfig(path string) *Config {
	var k = koanf.New(".")
	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		zap.L().Fatal("failed to load config", zap.Error(err))
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		zap.L().Fatal("failed to unmarshal config", zap.Error(err))
	}

	return &cfg
}

func (c *Config) GetDatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.DB.User, c.DB.Password, c.DB.Host, c.DB.Port, c.DB.DB, c.DB.SSLMode)
}

func (c *Config) GetServerAddr() string {
	return ":" + c.Server.Port
}
