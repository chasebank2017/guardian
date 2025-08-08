package config

import "github.com/spf13/viper"

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Auth     AuthConfig     `mapstructure:"auth"`
}

type AuthConfig struct {
	JWTSecret string `mapstructure:"jwt_secret"`
}

type ServerConfig struct {
	Port     string `mapstructure:"port"`
	GrpcPort string `mapstructure:"grpc_port"`
}

type DatabaseConfig struct {
	DSN string `mapstructure:"dsn"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./backend") // 在 backend 目录寻找 config.yaml
	viper.AddConfigPath("./")        // 兼容直接在根目录运行
	viper.AutomaticEnv()

	// Bind environment variables
	vip_err := viper.BindEnv("database.dsn", "DATABASE_URL")
	if vip_err != nil {
		return nil, vip_err
	}

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
