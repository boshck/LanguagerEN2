package handler

import (
	"go.uber.org/zap"
	tele "gopkg.in/telebot.v3"
)

// handleStart handles /start command
func (h *Handler) handleStart(c tele.Context) error {
	userID := c.Sender().ID

	h.logger.Info("User started bot",
		zap.Int64("user_id", userID),
		zap.String("username", c.Sender().Username),
	)

	// Ensure user exists in database
	if err := h.authService.EnsureUserExists(userID); err != nil {
		h.logger.Error("Failed to ensure user exists", zap.Error(err))
		return c.Send("Произошла ошибка. Попробуйте позже.")
	}

	// Check if authorized
	authorized, err := h.authService.IsAuthorized(userID)
	if err != nil {
		h.logger.Error("Failed to check authorization", zap.Error(err))
		return c.Send("Произошла ошибка. Попробуйте позже.")
	}

	if !authorized {
		// Request password
		h.ResetState(userID)
		return c.Send("Привет! Если ты не знаешь пароль, поздравляю - ты пукал, а коль знаешь - вводи:")
	}

	// Show main menu
	h.ResetState(userID)
	return c.Send(
		"🏠 Главное меню\n\nВыберите действие:",
		mainMenuMarkup(),
	)
}

