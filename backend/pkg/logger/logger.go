package logger

import (
	"log/slog"
	"os"
)

// Setup 初始化全局 JSON 格式 logger
func Setup() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(log)
}
