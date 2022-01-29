package logging

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *logger

type logger struct {
	*zap.Logger
}

func (l *logger) S() *zap.SugaredLogger {
	return l.Logger.Sugar()

}

func init() {
	pe := zap.NewProductionEncoderConfig()
	pe.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(pe)
	core := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zap.InfoLevel)

	l := zap.New(core)
	Log = &logger{
		Logger: l,
	}
}
