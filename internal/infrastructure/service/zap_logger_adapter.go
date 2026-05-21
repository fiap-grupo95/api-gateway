package service

import (
	"go.uber.org/zap"

	domsvc "github.com/fiap/secure-systems/api-gateway/internal/domain/service"
)

// ZapLoggerAdapter adapter que implementa a interface Logger usando zap
type ZapLoggerAdapter struct {
	logger *zap.Logger
}

// NewZapLoggerAdapter cria um novo adapter de logger usando zap
func NewZapLoggerAdapter(zapLogger *zap.Logger) domsvc.Logger {
	return &ZapLoggerAdapter{
		logger: zapLogger,
	}
}

func (z *ZapLoggerAdapter) Debug(msg string, keysAndValues ...interface{}) {
	z.logger.Sugar().Debugw(msg, keysAndValues...)
}

func (z *ZapLoggerAdapter) Info(msg string, keysAndValues ...interface{}) {
	z.logger.Sugar().Infow(msg, keysAndValues...)
}

func (z *ZapLoggerAdapter) Warn(msg string, keysAndValues ...interface{}) {
	z.logger.Sugar().Warnw(msg, keysAndValues...)
}

func (z *ZapLoggerAdapter) Error(msg string, keysAndValues ...interface{}) {
	z.logger.Sugar().Errorw(msg, keysAndValues...)
}

func (z *ZapLoggerAdapter) Fatal(msg string, keysAndValues ...interface{}) {
	z.logger.Sugar().Fatalw(msg, keysAndValues...)
}
