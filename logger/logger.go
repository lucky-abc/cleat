package logger

import (
	"github.com/natefinch/lumberjack"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var logger *zap.Logger
var config *viper.Viper
var logPath string

func NewLogger(lpath string, c *viper.Viper) {
	logPath = lpath
	config = c
	writeSyncer := getLogWriter()
	encoder := getEncoder()
	logLevel := zap.DebugLevel
	switch strings.ToLower(config.GetString("log.logLevel")) {
	case "debug":
		logLevel = zap.DebugLevel
	case "info":
		logLevel = zap.InfoLevel
	case "warn":
		logLevel = zap.WarnLevel
	case "error":
		logLevel = zap.ErrorLevel
	case "panic":
		logLevel = zap.PanicLevel
	case "fatal":
		logLevel = zap.FatalLevel
	default:
		logLevel = zap.InfoLevel
	}
	core := zapcore.NewCore(encoder, writeSyncer, logLevel)

	logger = zap.New(core, zap.AddCaller())
}

func Logger() *zap.Logger {
	return logger
}

func Loggers() *zap.SugaredLogger {
	return Logger().Sugar()
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter() zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filepath.Join(logPath, config.GetString("log.logFile.Filename")),
		MaxSize:    config.GetInt("log.logFile.MaxSize"),
		MaxBackups: config.GetInt("log.logFile.MaxBackups"),
		MaxAge:     config.GetInt("log.logFile.MaxAge"),
		Compress:   false,
	}
	return zapcore.NewMultiWriteSyncer(zapcore.AddSync(lumberJackLogger), zapcore.AddSync(os.Stdout))
}
