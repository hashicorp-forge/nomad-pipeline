package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewZap(cfg *Config) (*zap.Logger, error) {

	lvl, err := zap.ParseAtomicLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	enc := "console"
	ts := zapcore.ISO8601TimeEncoder

	if *cfg.JSON {
		enc = "json"
		ts = zapcore.RFC3339NanoTimeEncoder
	}

	baseCfg := zap.NewProductionConfig()
	baseCfg.DisableStacktrace = true
	baseCfg.Level = lvl
	baseCfg.Encoding = enc
	baseCfg.DisableCaller = !*cfg.IncludeLine
	baseCfg.EncoderConfig.NameKey = "component"
	baseCfg.EncoderConfig.TimeKey = "timestamp"
	baseCfg.EncoderConfig.EncodeTime = ts

	return baseCfg.Build()
}
