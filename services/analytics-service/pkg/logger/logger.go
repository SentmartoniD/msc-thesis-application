package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Log *zap.Logger
)

func LoadLoggerConfig() {
	Log, _ = Setup()

	defer Log.Sync()
}

// configure will return instance of zap logger configuration, configured to be verbose or to use JSON formatting
func Setup() (logger *zap.Logger, err error) {
	verbose := os.Getenv("VERBOSE")

	debugLevel := zapcore.InfoLevel
	switch verbose {
	case "debug":
		debugLevel = zapcore.DebugLevel
	case "info":
		debugLevel = zapcore.InfoLevel
	case "warn":
		debugLevel = zapcore.WarnLevel
	case "error":
		debugLevel = zapcore.ErrorLevel
	default:
		debugLevel = zapcore.InfoLevel
	}

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(debugLevel),
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          "console",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:     "message",
			LevelKey:       "level",
			TimeKey:        "time",
			NameKey:        "logger",
			CallerKey:      "go",
			StacktraceKey:  "trace",
			LineEnding:     "\n",
			EncodeLevel:    zapcore.CapitalColorLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller: func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
				callerName := caller.TrimmedPath()
				//callerName = filters.MinWidth(callerName, " ", 20)
				enc.AppendString(callerName)
			},
			EncodeName: zapcore.FullNameEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: nil,
		InitialFields:    nil,
	}

	return config.Build()
}
