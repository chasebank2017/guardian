package config

import "github.com/spf13/viper"

   Server   ServerConfig
   Database DatabaseConfig
   Auth     AuthConfig
}
type AuthConfig struct {
   JWTSecret string `mapstructure:"jwt_secret"`
}

type ServerConfig struct {
	Port     string
	GrpcPort string `mapstructure:"grpc_port"`
}

type DatabaseConfig struct {
	DSN string
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./backend") // 在 backend 目录寻找 config.yaml
	viper.AddConfigPath("./")        // 兼容直接在根目录运行
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
