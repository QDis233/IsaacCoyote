package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

func ApplyNewLogger(isDebug bool) (*zap.Logger, error) {
	encoderConfig := zapcore.EncoderConfig{
		MessageKey: "msg",
		LevelKey:   "level",
		TimeKey:    "time",

		NameKey:       "",
		CallerKey:     "",
		FunctionKey:   "",
		StacktraceKey: "",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.CapitalLevelEncoder,
		EncodeTime:    zapcore.RFC3339TimeEncoder,
	}
	level := zap.NewAtomicLevelAt(zapcore.InfoLevel)

	if isDebug {
		encoderConfig = zapcore.EncoderConfig{
			MessageKey: "msg",
			LevelKey:   "level",
			TimeKey:    "time",

			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    "func",
			StacktraceKey:  "stack",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.RFC3339TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.FullCallerEncoder,
			EncodeName:     zapcore.FullNameEncoder,
		}
		level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	}
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), os.Stdout, level)
	wrapOption := zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return core
	})

	zapConfig := zap.Config{}

	if isDebug {
		zapConfig = zap.NewDevelopmentConfig()
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	zapLogger, err := zapConfig.Build(
		zap.WithCaller(isDebug),
		zap.AddStacktrace(level),
		wrapOption,
	)
	if err != nil {
		zap.L().Error("failed to build zap logger", zap.Error(err))
		return nil, err
	}

	zap.ReplaceGlobals(zapLogger)
	return zapLogger, nil
}
