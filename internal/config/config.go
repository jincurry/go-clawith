package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	LLM      LLMConfig
}

type ServerConfig struct {
	Port string
	Mode string // debug, release
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret     string
	ExpireHour int
}

type LLMConfig struct {
	DefaultProvider string
	OpenAI          OpenAIConfig
	Anthropic       AnthropicConfig
}

type OpenAIConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

type AnthropicConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	viper.AutomaticEnv()

	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "clawith")
	viper.SetDefault("database.password", "clawith")
	viper.SetDefault("database.dbname", "clawith")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("redis.addr", "localhost:6379")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("jwt.secret", "change-me-in-production")
	viper.SetDefault("jwt.expire_hour", 72)
	viper.SetDefault("llm.default_provider", "openai")
	viper.SetDefault("llm.openai.base_url", "https://api.openai.com/v1")
	viper.SetDefault("llm.openai.model", "gpt-4o")
	viper.SetDefault("llm.anthropic.base_url", "https://api.anthropic.com")
	viper.SetDefault("llm.anthropic.model", "claude-sonnet-4-20250514")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	cfg := &Config{
		Server: ServerConfig{
			Port: viper.GetString("server.port"),
			Mode: viper.GetString("server.mode"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("database.host"),
			Port:     viper.GetInt("database.port"),
			User:     viper.GetString("database.user"),
			Password: viper.GetString("database.password"),
			DBName:   viper.GetString("database.dbname"),
			SSLMode:  viper.GetString("database.sslmode"),
		},
		Redis: RedisConfig{
			Addr:     viper.GetString("redis.addr"),
			Password: viper.GetString("redis.password"),
			DB:       viper.GetInt("redis.db"),
		},
		JWT: JWTConfig{
			Secret:     viper.GetString("jwt.secret"),
			ExpireHour: viper.GetInt("jwt.expire_hour"),
		},
		LLM: LLMConfig{
			DefaultProvider: viper.GetString("llm.default_provider"),
			OpenAI: OpenAIConfig{
				APIKey:  viper.GetString("llm.openai.api_key"),
				BaseURL: viper.GetString("llm.openai.base_url"),
				Model:   viper.GetString("llm.openai.model"),
			},
			Anthropic: AnthropicConfig{
				APIKey:  viper.GetString("llm.anthropic.api_key"),
				BaseURL: viper.GetString("llm.anthropic.base_url"),
				Model:   viper.GetString("llm.anthropic.model"),
			},
		},
	}

	return cfg, nil
}
