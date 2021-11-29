package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

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
}

var Hierarchical = map[string]zapcore.Level{
	"info": zapcore.InfoLevel,
	"warn": zapcore.WarnLevel,
	"error": zapcore.ErrorLevel,
	"dpanic": zapcore.DPanicLevel,
	"panic": zapcore.PanicLevel,
	//"fatal": zapcore.FatalLevel,
}


type Log struct {
	Config Config
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

func (l *Log) InitLog () *zap.Logger{
	l.GetEncoder()

	if !l.Config.Hierarchical {
		core := zapcore.NewCore(zapcore.NewJSONEncoder(l.Encoder),
			zapcore.NewMultiWriteSyncer(l.GetLogWriter(), os.Stdout), zap.DebugLevel)

		return zap.New(core, zap.AddStacktrace(zap.DebugLevel))
	}


	cores := make([]zapcore.Core, 0)
	for level, hierarchical := range Hierarchical{
		l.Config.Filename = level + "/" + l.Config.Filename
		core := zapcore.NewCore(zapcore.NewJSONEncoder(l.Encoder), l.GetLogWriter(), zap.LevelEnablerFunc(func(level zapcore.Level) bool {
			return level == hierarchical
		}))
		cores = append(cores, core)
	}

	cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(l.Encoder), zapcore.AddSync(os.Stdout), zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level == zapcore.FatalLevel
	})))

	cores = append(cores, zapcore.NewCore(zapcore.NewJSONEncoder(l.Encoder),
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)), zapcore.DebugLevel))

	core := zapcore.NewTee(cores...)

	if l.Config.Debug {
		return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.DebugLevel))
	}

	return zap.New(core)
}

func(l *Log) GetEncoder() {
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

func(l *Log) GetLogWriter() zapcore.WriteSyncer {
	hook := Logger{
		Dir:        l.Config.Dir,
		Filename:   l.Config.Filename,
		ServerName: l.Config.ServerName,
		MaxSize:    l.Config.MaxSize,
		MaxBackups: l.Config.MaxBackups,
		MaxAge:     l.Config.MaxAge,
		Compress:   l.Config.Compress,
	}

	return zapcore.AddSync(&hook)
}
