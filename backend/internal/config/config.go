package config

import "github.com/spf13/viper"

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Auth     AuthConfig     `mapstructure:"auth"`
}

type AuthConfig struct {
	JWTSecret string `mapstructure:"jwt_secret"`
    AdminUsername string `mapstructure:"admin_username"`
    AdminPassword string `mapstructure:"admin_password"`
}

type ServerConfig struct {
	Port     string `mapstructure:"port"`
	GrpcPort string `mapstructure:"grpc_port"`
    RequestTimeoutSeconds int `mapstructure:"request_timeout_seconds"`
    CORSOrigins []string `mapstructure:"cors_origins"`
    RateLimit   RateLimitConfig `mapstructure:"rate_limit"`
}

type RateLimitConfig struct {
    LoginRPS     int `mapstructure:"login_rps"`
    LoginBurst   int `mapstructure:"login_burst"`
    ProtectedRPS int `mapstructure:"protected_rps"`
    ProtectedBurst int `mapstructure:"protected_burst"`
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
    // Optional env override for admin credentials
    _ = viper.BindEnv("auth.admin_username", "ADMIN_USERNAME")
    _ = viper.BindEnv("auth.admin_password", "ADMIN_PASSWORD")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
