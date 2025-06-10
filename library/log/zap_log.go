package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"strings"
)

const (
	logLevel  = "log_level"
	logOutput = "log_output"
)

var (
	log Logger
)

type AppLoggerInterface interface {
	Graylog(bool) AppLoggerInterface
	Info(string, ...zap.Field)
	Error(string, error, ...zap.Field)
	Debug(string, ...zap.Field)
	Warning(string, error, ...zap.Field)
	Panic(string, error, ...zap.Field)
	Fatal(string, error, ...zap.Field)
}

type Logger struct {
	log           *zap.Logger
	sendToGraylog bool
}

func init() {
	logConfig := zap.Config{
		OutputPaths: []string{getOutput()},
		Level:       zap.NewAtomicLevelAt(getLevel()),
		Encoding:    "json",
		EncoderConfig: zapcore.EncoderConfig{
			LevelKey:     "level",
			TimeKey:      "time",
			MessageKey:   "message",
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			EncodeLevel:  zapcore.LowercaseLevelEncoder,
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}

	var err error
	if log.log, err = logConfig.Build(); err != nil {
		panic(err)
	}
}

func getOutput() string {
	output := strings.ToLower(strings.TrimSpace(os.Getenv(logOutput)))
	if output == "" {
		return "stdout"
	}

	return output
}

func getLevel() zapcore.Level {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(logLevel))) {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "error":
		return zap.ErrorLevel
	case "warning":
		return zap.WarnLevel
	case "panic":
		return zap.PanicLevel
	case "fatal":
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}

func GetLogger() AppLoggerInterface {
	log.Graylog(false)

	return log
}

func (l Logger) Graylog(sendToGraylog bool) AppLoggerInterface {
	log.sendToGraylog = sendToGraylog

	return log
}

func (l Logger) Info(message string, v ...zap.Field) {
	InfoL(message, v...)
}

func (l Logger) Error(message string, err error, v ...zap.Field) {
	v = append(v, zap.String("error", err.Error()))
	ErrorL(message, v...)
}

func (l Logger) Debug(message string, v ...zap.Field) {
	DebugL(message, v...)
}

func (l Logger) Warning(message string, err error, v ...zap.Field) {
	v = append(v, zap.String("error", err.Error()))
	WarningL(message, v...)
}

func (l Logger) Panic(message string, err error, v ...zap.Field) {
	v = append(v, zap.String("error", err.Error()))
	PanicL(message, v...)
}

func (l Logger) Fatal(message string, err error, v ...zap.Field) {
	v = append(v, zap.String("error", err.Error()))
	FatalL(message, v...)
}

func InfoL(message string, tags ...zap.Field) {
	log.log.Info(message, tags...)
	_ = log.log.Sync()
}

func ErrorL(message string, tags ...zap.Field) {
	log.log.Error(message, tags...)
	_ = log.log.Sync()
}

func DebugL(message string, tags ...zap.Field) {
	log.log.Debug(message, tags...)
	_ = log.log.Sync()
}

func WarningL(message string, tags ...zap.Field) {
	log.log.Warn(message, tags...)
	_ = log.log.Sync()
}

func PanicL(message string, tags ...zap.Field) {
	log.log.Panic(message, tags...)
	_ = log.log.Sync()
}

func FatalL(message string, tags ...zap.Field) {
	log.log.Fatal(message, tags...)
	_ = log.log.Sync()
}
