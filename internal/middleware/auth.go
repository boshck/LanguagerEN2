package middleware

import (
	"languager/internal/service"

	"go.uber.org/zap"
	tele "gopkg.in/telebot.v3"
)

// AuthMiddleware creates authentication middleware
func AuthMiddleware(authService *service.AuthService, logger *zap.Logger) tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			userID := c.Sender().ID

			// Ensure user exists
			if err := authService.EnsureUserExists(userID); err != nil {
				logger.Error("Failed to ensure user exists in middleware", zap.Error(err))
				return c.Send("Произошла ошибка. Попробуйте позже.")
			}

			// Check authorization
			authorized, err := authService.IsAuthorized(userID)
			if err != nil {
				logger.Error("Failed to check authorization in middleware", zap.Error(err))
				return c.Send("Произошла ошибка. Попробуйте позже.")
			}

			// If not authorized and not /start command, prompt for password
			if !authorized && c.Text() != "/start" {
				return c.Send("Привет! Если ты не знаешь пароль, поздравляю - ты пукал, а коль знаешь - вводи:")
			}

			// User is authorized or using /start, continue
			return next(c)
		}
	}
}

