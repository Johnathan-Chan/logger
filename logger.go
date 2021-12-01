package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var Hierarchical = map[string][2]zapcore.Level{
	"info": {zapcore.WarnLevel, zapcore.DebugLevel},
	"warn": {zapcore.ErrorLevel, zapcore.InfoLevel},
	"error": {zapcore.DPanicLevel, zapcore.WarnLevel},
	"painc": {zapcore.FatalLevel, zapcore.DPanicLevel},
}

type Config struct {
	Dir          string
	Filename     string
	ServerName   string
	Hierarchical bool
	Debug        bool
	MaxSize      int
	MaxBackups   int
	MaxAge       int
	Compress     bool
	Console      bool
}

type Log struct {
	Config  Config
	Encoder zapcore.EncoderConfig
	*zap.Logger
}

func NewLogger(config Config) *Log {
	log := &Log{
		Config: config,
	}

	log.Logger = log.InitLog()
	return log
}

func (l *Log) InitLog() *zap.Logger {
	l.GetEncoder()

	if !l.Config.Hierarchical {
		cores := []zapcore.Core{
			zapcore.NewCore(zapcore.NewJSONEncoder(l.Encoder), l.GetLogWriter(""), zap.DebugLevel),
		}

		if l.Config.Console {
			cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(l.Encoder), zapcore.AddSync(os.Stdout), zap.DebugLevel))
		}

		core := zapcore.NewTee(cores...)


		if l.Config.Debug{
			return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.DebugLevel))
		}

		return zap.New(core)
	}

	cores := make([]zapcore.Core, 0)
	cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(l.Encoder), l.GetLogWriter("info"+"-"), zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.WarnLevel && lvl >= zapcore.DebugLevel
	})))

	cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(l.Encoder), l.GetLogWriter("warn"+"-"), zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel && lvl >= zapcore.WarnLevel
	})))

	cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(l.Encoder), l.GetLogWriter("error"+"-"), zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.DPanicLevel && lvl >= zapcore.ErrorLevel
	})))

	cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(l.Encoder), l.GetLogWriter("painc"+"-"), zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.DPanicLevel
	})))

	cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(l.Encoder), zapcore.AddSync(os.Stdout), zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level == zapcore.FatalLevel
	})))

	if l.Config.Console {
		cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(l.Encoder),
			zapcore.AddSync(os.Stdout), zapcore.DebugLevel))
	}

	core := zapcore.NewTee(cores...)

	if l.Config.Debug {
		return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.DebugLevel))
	}

	return zap.New(core)
}

func (l *Log) GetEncoder() {
	l.Encoder = zap.NewProductionEncoderConfig()
	l.Encoder.TimeKey = "timestamp"
	l.Encoder.LevelKey = "level"
	l.Encoder.NameKey = "logger"
	l.Encoder.CallerKey = "caller"
	l.Encoder.MessageKey = "message"
	l.Encoder.StacktraceKey = "stacktrace"
	l.Encoder.EncodeCaller = zapcore.ShortCallerEncoder
	l.Encoder.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	l.Encoder.EncodeLevel = zapcore.CapitalLevelEncoder
}

func (l *Log) GetLogWriter(prefix string) zapcore.WriteSyncer {
	hook := Logger{
		Dir:        l.Config.Dir,
		Filename:   prefix+l.Config.Filename,
		ServerName: l.Config.ServerName,
		MaxSize:    l.Config.MaxSize,
		MaxBackups: l.Config.MaxBackups,
		MaxAge:     l.Config.MaxAge,
		Compress:   l.Config.Compress,
	}

	return zapcore.AddSync(&hook)
}
