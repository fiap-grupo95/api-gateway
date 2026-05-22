package service

import (
	"github.com/fiap/secure-systems/api-gateway/internal/logging"
	domsvc "github.com/fiap/secure-systems/api-gateway/internal/domain/service"
)

type ZerologLoggerAdapter struct{}

func NewZerologLoggerAdapter() domsvc.Logger {
	return &ZerologLoggerAdapter{}
}

func (z *ZerologLoggerAdapter) Debug(msg string, kv ...interface{}) {
	logging.Logger().Debug().Fields(kv).Msg(msg)
}

func (z *ZerologLoggerAdapter) Info(msg string, kv ...interface{}) {
	logging.Logger().Info().Fields(kv).Msg(msg)
}

func (z *ZerologLoggerAdapter) Warn(msg string, kv ...interface{}) {
	logging.Logger().Warn().Fields(kv).Msg(msg)
}

func (z *ZerologLoggerAdapter) Error(msg string, kv ...interface{}) {
	logging.Logger().Error().Fields(kv).Msg(msg)
}

func (z *ZerologLoggerAdapter) Fatal(msg string, kv ...interface{}) {
	logging.Logger().Fatal().Fields(kv).Msg(msg)
}
