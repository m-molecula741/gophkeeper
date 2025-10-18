// logger/logger.go
package logger

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

func New(level string) (*zap.Logger, error) {
    var lvl zapcore.Level
    switch level {
    case "debug":
        lvl = zap.DebugLevel
    case "info":
        lvl = zap.InfoLevel
    case "warn":
        lvl = zap.WarnLevel
    case "error":
        lvl = zap.ErrorLevel
    case "fatal":
        lvl = zap.FatalLevel
    default:
        lvl = zap.InfoLevel
    }

    cfg := zap.NewProductionConfig()
    cfg.Level = zap.NewAtomicLevelAt(lvl)
    cfg.OutputPaths = []string{"stdout"}
    cfg.ErrorOutputPaths = []string{"stderr"}

    return cfg.Build()
}