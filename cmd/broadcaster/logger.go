package main

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func buildZapLogger(encoding string) (*zap.Logger, error) {
	var zapLogger *zap.Logger
	var err error

	if encoding == "json" {
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.MessageKey = "message"
		encoderConfig.LevelKey = "severity"
		encoderConfig.TimeKey = "timestamp"
		encoderConfig.NameKey = "logger"
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		encoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder

		config := zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		config.EncoderConfig = encoderConfig

		zapLogger, err = config.Build(
			zap.AddCallerSkip(1),
		)
	} else {
		encoderConfig := zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

		config := zap.NewDevelopmentConfig()
		config.EncoderConfig = encoderConfig

		zapLogger, err = config.Build(
			zap.AddCallerSkip(1),
		)
	}

	return zapLogger, err
}
