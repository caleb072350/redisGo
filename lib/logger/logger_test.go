package logger_test

import (
	"redisGo/lib/logger"
	"testing"
)

func TestLogger(t *testing.T) {
	settings := &logger.Settings{
		Path:       "./logs",
		Name:       "myGodis",
		Ext:        "log",
		TimeFormat: "2006-01-02",
	}
	logger.Setup(settings)

	// 测试日志记录
	logger.Debug("This is a debug message")
	logger.Info("This is an info message")
	logger.Warn("This is a warn message")
	logger.Error("This is an error message")
	logger.Fatal("This is a fatal message")
}
