package utils

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerType int

const (
	DebugMode LoggerType = iota
	DeploymentMode
	ConsoleLogger
	ServerLogger
)

// InitLogger ...
func InitLogger(mode, lType LoggerType) *zap.Logger {

	// TimeKey:        "T",
	// LevelKey:       "L",
	// NameKey:        "N",
	// CallerKey:      "C",
	// MessageKey:     "M",
	// StacktraceKey: "S",

	var level zap.AtomicLevel

	enCfg := zap.NewDevelopmentEncoderConfig()
	enCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	if mode == DeploymentMode {
		enCfg.TimeKey = "T"
		level = zap.NewAtomicLevelAt(zap.InfoLevel)
	} else {
		enCfg.TimeKey = ""
		level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	if lType == ConsoleLogger {
		enCfg.TimeKey = ""
		enCfg.LevelKey = "L"
		enCfg.NameKey = ""
		enCfg.CallerKey = ""
	}

	if mode == DebugMode && lType == ConsoleLogger {
		enCfg.NameKey = "N"
		enCfg.CallerKey = "C"
	}

	cfg := zap.Config{
		Level:             level,
		Development:       true,
		Encoding:          "console",
		EncoderConfig:     enCfg,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
		DisableStacktrace: true,
	}
	logger, err := cfg.Build()
	if err != nil {
		fmt.Println("Error: ", err)
		return nil
	}

	zap.ReplaceGlobals(logger)

	return logger
}
