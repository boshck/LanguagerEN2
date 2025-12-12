package handler

import (
	"strings"

	"languager/internal/domain"

	"go.uber.org/zap"
	tele "gopkg.in/telebot.v3"
)

// handleText handles all text messages based on state
func (h *Handler) handleText(c tele.Context) error {
	userID := c.Sender().ID
	text := strings.TrimSpace(c.Text())

	// Ignore commands (starting with /)
	if strings.HasPrefix(text, "/") {
		return nil
	}

	// Ensure user exists
	if err := h.authService.EnsureUserExists(userID); err != nil {
		h.logger.Error("Failed to ensure user exists", zap.Error(err))
		return nil
	}

	// Check authorization first
	authorized, err := h.authService.IsAuthorized(userID)
	if err != nil {
		h.logger.Error("Failed to check authorization", zap.Error(err))
		return c.Send("Произошла ошибка. Попробуйте позже.")
	}

	// If not authorized, check password
	if !authorized {
		if h.authService.CheckPassword(text) {
			// Correct password
			if err := h.authService.AuthorizeUser(userID); err != nil {
				h.logger.Error("Failed to authorize user", zap.Error(err))
				return c.Send("Произошла ошибка. Попробуйте позже.")
			}

			h.logger.Info("User authorized", zap.Int64("user_id", userID))
			h.ResetState(userID)
			return c.Send(
				"✅ Доступ разрешён!\n\n🏠 Главное меню\n\nВыберите действие:",
				mainMenuMarkup(),
			)
		}

		// Wrong password
		return c.Send("Непральна")
	}

	// User is authorized, handle based on state
	state := h.GetState(userID)

	switch state.State {
	case domain.StateWaitingWord:
		// User sent a word, now wait for translation
		cancelMarkup := &tele.ReplyMarkup{}
		cancelMarkup.Inline(cancelMarkup.Row(btnCancel))

		h.SetState(userID, &domain.StateData{
			State:       domain.StateWaitingTranslation,
			CurrentWord: text,
		})

		return c.Send("Жду перевод", cancelMarkup)

	case domain.StateWaitingTranslation:
		// User sent translation, save the pair
		word := state.CurrentWord
		translation := text

		if err := h.wordService.SaveWordPair(userID, word, translation); err != nil {
			h.logger.Error("Failed to save word pair",
				zap.Error(err),
				zap.Int64("user_id", userID),
			)
			return c.Send("Не удалось сохранить слово. Попробуйте ещё раз.")
		}

		h.logger.Info("Word pair saved",
			zap.Int64("user_id", userID),
			zap.String("word", word),
			zap.String("translation", translation),
		)

		// Reset to waiting for next word
		h.SetState(userID, &domain.StateData{State: domain.StateWaitingWord})

		return c.Send("✅ Сохранено!\n\nМожешь отправить следующее слово или вернуться в /start")

	default:
		// Idle state - start word input flow
		cancelMarkup := &tele.ReplyMarkup{}
		cancelMarkup.Inline(cancelMarkup.Row(btnCancel))

		h.SetState(userID, &domain.StateData{
			State:       domain.StateWaitingTranslation,
			CurrentWord: text,
		})

		return c.Send("Жду перевод", cancelMarkup)
	}
}

