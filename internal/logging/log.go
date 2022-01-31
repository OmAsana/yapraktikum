package logging

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
	atomicLevel *zap.AtomicLevel
}

func (l *Logger) Flush() {
	l.Logger.Sync()
}

func (l *Logger) S() *zap.SugaredLogger {
	return l.Logger.Sugar()

}

func (l *Logger) SetLogLevel(level zapcore.Level) {
	l.atomicLevel.SetLevel(level)
}

func NewLogger() *Logger {
	pe := zap.NewProductionEncoderConfig()
	pe.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(pe)

	logLevel := zap.NewAtomicLevel()

	core := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), logLevel)

	l := zap.New(core)
	return &Logger{
		Logger:      l,
		atomicLevel: &logLevel,
	}
}

func NewNoop() *Logger {
	return &Logger{
		Logger: zap.NewNop(),
	}

}
