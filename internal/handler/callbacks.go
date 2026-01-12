package handler

import (
	"fmt"
	"html"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"go.uber.org/zap"
	tele "gopkg.in/telebot.v3"
)

// cleanCallbackData removes all non-printable characters from callback data
func cleanCallbackData(data string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, strings.TrimSpace(data))
}

// handleEditError handles errors from c.Edit() - просто логирует ошибки
// Callback уже должен быть подтверждён до вызова этой функции
func (h *Handler) handleEditError(err error, c tele.Context, userID int64) bool {
	if err == nil {
		return false
	}
	
	errStr := err.Error()
	// If message is not modified, it means it was already edited by another callback
	if strings.Contains(errStr, "message is not modified") {
		h.logger.Debug("Message already modified by another callback",
			zap.Int64("user_id", userID),
			zap.String("callback_id", c.Callback().ID),
		)
		return true // Уже изменено, это нормально
	}
	
	// Log the error
	h.logger.Warn("Failed to edit message",
		zap.Error(err),
		zap.Int64("user_id", userID),
		zap.String("callback_id", c.Callback().ID),
	)
	return false // Реальная ошибка
}

// handleCallback handles ALL callback queries
func (h *Handler) handleCallback(c tele.Context) error {
	callback := c.Callback()
	if callback == nil {
		h.logger.Warn("handleCallback: callback is nil")
		return nil
	}

	// Clean data from all non-printable characters
	data := cleanCallbackData(callback.Data)
	h.logger.Info("handleCallback: Processing callback",
		zap.String("data", data),
		zap.String("data_raw", callback.Data),
		zap.String("id", callback.ID),
		zap.String("unique", callback.Unique),
		zap.Int64("user_id", c.Sender().ID),
	)

	// Handle specific button callbacks by Unique first
	switch callback.Unique {
	case "view_days", "back_to_days":
		return h.handleViewDays(c)
	case "random_pair", "more":
		return h.handleRandomPair(c)
	case "cancel":
		return h.handleCancel(c)
	case "back", "main_menu":
		return h.handleStart(c)
	}

	// If Unique is empty, try to handle by Data (for buttons with Unique that didn't come through)
	if callback.Unique == "" {
		switch data {
		case "view_days", "back_to_days":
			return h.handleViewDays(c)
		case "random_pair", "more":
			return h.handleRandomPair(c)
		case "cancel":
			return h.handleCancel(c)
		case "back", "main_menu":
			return h.handleStart(c)
		}
	}

	// Handle by Data prefix (dynamic buttons)
	switch {
	case strings.HasPrefix(data, "page_"):
		return h.handlePagination(c, data)
	case strings.HasPrefix(data, "day_"):
		return h.handleDaySelection(c, data)
	case strings.HasPrefix(data, "hide_7d_"):
		return h.handleHideFor7Days(c, data)
	case strings.HasPrefix(data, "hide_forever_"):
		return h.handleHideForeverConfirm(c, data)
	case strings.HasPrefix(data, "confirm_hide_"):
		return h.handleConfirmHideForever(c, data)
	case strings.HasPrefix(data, "cancel_hide_"):
		return h.handleCancelHide(c, data)
	}

	// If it's not handled, acknowledge it anyway
	h.logger.Warn("Unhandled callback in handleCallback",
		zap.String("data", data),
		zap.String("unique", callback.Unique),
	)
	return c.Respond()
}

// handleViewDays shows list of days with words
func (h *Handler) handleViewDays(c tele.Context) error {
	userID := c.Sender().ID

	// КРИТИЧЕСКИ ВАЖНО: Отвечаем на callback СРАЗУ
	if c.Callback() != nil {
		if err := c.Respond(); err != nil {
			h.logger.Warn("Failed to acknowledge callback", zap.Error(err))
		}
	}

	// Get first page
	days, totalPages, err := h.wordService.GetDaysList(userID, 1)
	if err != nil {
		h.logger.Error("Failed to get days list", zap.Error(err))
		return nil // Callback уже подтверждён
	}

	if len(days) == 0 {
		// Callback уже подтверждён, ничего не делаем
		return nil
	}

	// Build message
	text := "📅 Вот твои дни:\n\n"
	markup := &tele.ReplyMarkup{}
	rows := []tele.Row{}

	for _, day := range days {
		btnText := fmt.Sprintf("%s (%d)", day.DisplayString(), day.WordCount)
		btn := markup.Data(btnText, "day_"+day.DateString())
		rows = append(rows, markup.Row(btn))
	}

	// Add pagination buttons if needed
	if totalPages > 1 {
		navRow := tele.Row{}
		// First page, only show "Next"
		navRow = append(navRow, markup.Data("➡️", "page_2"))
		rows = append(rows, navRow)
	}

	// Add back button
	rows = append(rows, markup.Row(btnBack))

	markup.Inline(rows...)

	// Edit message - только edit, никаких send
	if c.Callback() != nil {
		if err := c.Edit(text, markup); err != nil {
			h.handleEditError(err, c, userID)
			// Callback уже подтверждён, просто логируем ошибку
		}
		return nil
	}
	// Это не callback (например команда), можно отправлять новое
	return c.Send(text, markup)
}

// handleRandomPair shows a random word-translation pair
func (h *Handler) handleRandomPair(c tele.Context) error {
	userID := c.Sender().ID

	// КРИТИЧЕСКИ ВАЖНО: Отвечаем на callback СРАЗУ, до блокировки
	if c.Callback() != nil {
		if err := c.Respond(); err != nil {
			h.logger.Warn("Failed to acknowledge callback immediately", zap.Error(err))
		}
	}

	// Получаем или создаём блокировку для этого пользователя
	h.callbackMux.Lock()
	lock, exists := h.callbackLocks[userID]
	if !exists {
		lock = &sync.Mutex{}
		h.callbackLocks[userID] = lock
	}
	h.callbackMux.Unlock()

	// Блокируем обработку для этого пользователя
	lock.Lock()
	defer lock.Unlock()

	word, err := h.wordService.GetRandomPair(userID)
	if err != nil {
		h.logger.Error("Failed to get random word", zap.Error(err))
		return nil // Callback уже подтверждён
	}

	if word == nil {
		// Callback уже подтверждён
		return nil
	}

	// Рандомно выбираем, что показывать открыто, а что под спойлером
	rand.Seed(time.Now().UnixNano())
	showWordFirst := rand.Intn(2) == 0

	escWord := html.EscapeString(word.Word)
	escTranslation := html.EscapeString(word.Translation)

	var visibleText, spoilerText string
	if showWordFirst {
		visibleText = fmt.Sprintf("📝 %s", escWord)
		spoilerText = fmt.Sprintf("🔄 %s", escTranslation)
	} else {
		visibleText = fmt.Sprintf("🔄 %s", escTranslation)
		spoilerText = fmt.Sprintf("📝 %s", escWord)
	}

	// Формируем текст со спойлером в формате HTML
	// В Telegram Bot API спойлеры работают через тег <tg-spoiler>текст</tg-spoiler>
	text := fmt.Sprintf("🎲 Случайная пара:\n\n%s\n<tg-spoiler>%s</tg-spoiler>", visibleText, spoilerText)

	markup := &tele.ReplyMarkup{}
	markup.Inline(
		markup.Row(btnMore),
		markup.Row(
			markup.Data("💤 Не показывать 7 дней", fmt.Sprintf("hide_7d_%d", word.ID)),
			markup.Data("♿️ Не показывать никогда", fmt.Sprintf("hide_forever_%d", word.ID)),
		),
		markup.Row(btnBack),
	)

	// Edit message - только edit, никаких send
	// Указываем режим парсинга HTML для поддержки спойлеров
	// В Telegram Bot API можно одновременно использовать parse_mode и reply_markup
	// Ссылка: https://core.telegram.org/bots/api#editmessagetext
	// В telebot.v3 нужно использовать Bot.Edit() напрямую с опциями
	if c.Callback() != nil {
		opts := &tele.SendOptions{
			ParseMode:   "HTML",
			ReplyMarkup: markup,
		}
		if _, err := h.bot.Edit(c.Callback().Message, text, opts); err != nil {
			h.handleEditError(err, c, userID)
			// Callback уже подтверждён, просто логируем ошибку
		}
		return nil
	}
	// Это не callback (например команда), можно отправлять новое
	return c.Send(text, markup, &tele.SendOptions{ParseMode: "HTML"})
}

// handleCancel cancels current operation and resets state
func (h *Handler) handleCancel(c tele.Context) error {
	userID := c.Sender().ID

	// КРИТИЧЕСКИ ВАЖНО: Отвечаем на callback СРАЗУ
	if c.Callback() != nil {
		if err := c.Respond(); err != nil {
			h.logger.Warn("Failed to acknowledge callback", zap.Error(err))
		}
	}

	h.ResetState(userID)

	if err := c.Edit(
		"🏠 Главное меню\n\nВыберите действие:",
		mainMenuMarkup(),
	); err != nil {
		h.handleEditError(err, c, userID)
		// Callback уже подтверждён, просто логируем ошибку
	}
	return nil
}

// handlePagination handles page navigation
func (h *Handler) handlePagination(c tele.Context, data string) error {
	userID := c.Sender().ID

	// КРИТИЧЕСКИ ВАЖНО: Отвечаем на callback СРАЗУ
	if c.Callback() != nil {
		if err := c.Respond(); err != nil {
			h.logger.Warn("Failed to acknowledge callback", zap.Error(err))
		}
	}

	// Extract page number - trim whitespace first
	data = strings.TrimSpace(data)
	pageStr := strings.TrimPrefix(data, "page_")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		// Callback уже подтверждён, просто возвращаемся
		return nil
	}

	days, totalPages, err := h.wordService.GetDaysList(userID, page)
	if err != nil {
		h.logger.Error("Failed to get days list", zap.Error(err))
		return nil // Callback уже подтверждён
	}

	if len(days) == 0 {
		// Callback уже подтверждён
		return nil
	}

	// Build message
	text := "📅 Вот твои дни:\n\n"
	markup := &tele.ReplyMarkup{}
	rows := []tele.Row{}

	for _, day := range days {
		btnText := fmt.Sprintf("%s (%d)", day.DisplayString(), day.WordCount)
		btn := markup.Data(btnText, "day_"+day.DateString())
		rows = append(rows, markup.Row(btn))
	}

	// Add pagination buttons
	if totalPages > 1 {
		navRow := tele.Row{}
		if page > 1 {
			navRow = append(navRow, markup.Data("⬅️", fmt.Sprintf("page_%d", page-1)))
		}
		if page < totalPages {
			navRow = append(navRow, markup.Data("➡️", fmt.Sprintf("page_%d", page+1)))
		}
		if len(navRow) > 0 {
			rows = append(rows, navRow)
		}
	}

	// Add back button
	rows = append(rows, markup.Row(btnBack))

	markup.Inline(rows...)

	// Edit message - только edit, никаких send
	if err := c.Edit(text, markup); err != nil {
		h.handleEditError(err, c, userID)
		// Callback уже подтверждён, просто логируем ошибку
	}
	return nil
}

// handleDaySelection shows words for selected day
func (h *Handler) handleDaySelection(c tele.Context, data string) error {
	userID := c.Sender().ID

	// КРИТИЧЕСКИ ВАЖНО: Отвечаем на callback СРАЗУ
	if c.Callback() != nil {
		if err := c.Respond(); err != nil {
			h.logger.Warn("Failed to acknowledge callback", zap.Error(err))
		}
	}

	// Extract date - trim whitespace first, then remove prefix
	data = strings.TrimSpace(data)
	dateStr := strings.TrimPrefix(data, "day_")
	h.logger.Info("Handling day selection", zap.String("date", dateStr), zap.String("original_data", data), zap.Int64("user_id", userID))

	words, err := h.wordService.GetWordsByDate(userID, dateStr)
	if err != nil {
		h.logger.Error("Failed to get words by date", zap.Error(err))
		return nil // Callback уже подтверждён
	}

	if len(words) == 0 {
		// Callback уже подтверждён
		return nil
	}

	// Build message with all words
	text := fmt.Sprintf("📝 Слова за выбранный день (%d):\n\n", len(words))
	for i, word := range words {
		// Determine status emoji
		var statusEmoji string
		if word.HiddenForever {
			statusEmoji = "♿️"
		} else if word.HiddenUntil != nil && word.HiddenUntil.After(time.Now()) {
			statusEmoji = "💤"
		} else {
			statusEmoji = "💡"
		}
		text += fmt.Sprintf("%d. %s %s — %s\n\n", i+1, statusEmoji, word.Word, word.Translation)
	}

	markup := &tele.ReplyMarkup{}
	markup.Inline(
		markup.Row(btnBackToDays, btnMainMenu),
	)

	// Edit message - только edit, никаких send
	if err := c.Edit(text, markup); err != nil {
		h.handleEditError(err, c, userID)
		// Callback уже подтверждён, просто логируем ошибку
	}
	return nil
}

// handleHideFor7Days hides a word for 7 days and shows success message with "Ещё" button
func (h *Handler) handleHideFor7Days(c tele.Context, data string) error {
	userID := c.Sender().ID

	// КРИТИЧЕСКИ ВАЖНО: Отвечаем на callback СРАЗУ
	if c.Callback() != nil {
		if err := c.Respond(); err != nil {
			h.logger.Warn("Failed to acknowledge callback", zap.Error(err))
		}
	}

	// Extract word ID
	data = strings.TrimSpace(data)
	wordIDStr := strings.TrimPrefix(data, "hide_7d_")
	wordID, err := strconv.Atoi(wordIDStr)
	if err != nil {
		h.logger.Error("Failed to parse word ID", zap.Error(err), zap.String("data", data))
		return nil // Callback уже подтверждён
	}

	// Hide the word
	if err := h.wordService.HideWordFor7Days(wordID); err != nil {
		h.logger.Error("Failed to hide word for 7 days", zap.Error(err), zap.Int("word_id", wordID))
		return nil // Callback уже подтверждён
	}

	// Show success message with "Ещё" button
	text := "✅ Слово скрыто на 7 дней"
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(btnMore))

	// Edit message - только edit, никаких send
	if err := c.Edit(text, markup); err != nil {
		h.handleEditError(err, c, userID)
		// Callback уже подтверждён, просто логируем ошибку
	}
	return nil
}

// handleHideForeverConfirm shows confirmation dialog for permanent hiding
func (h *Handler) handleHideForeverConfirm(c tele.Context, data string) error {
	userID := c.Sender().ID

	// КРИТИЧЕСКИ ВАЖНО: Отвечаем на callback СРАЗУ
	if c.Callback() != nil {
		if err := c.Respond(); err != nil {
			h.logger.Warn("Failed to acknowledge callback", zap.Error(err))
		}
	}

	// Extract word ID
	data = strings.TrimSpace(data)
	wordIDStr := strings.TrimPrefix(data, "hide_forever_")
	wordID, err := strconv.Atoi(wordIDStr)
	if err != nil {
		h.logger.Error("Failed to parse word ID", zap.Error(err), zap.String("data", data))
		return nil // Callback уже подтверждён
	}

	// Show confirmation message
	text := "❓ Точно ли хочешь убрать слово из повторения? Его придётся внести ещё раз"
	markup := &tele.ReplyMarkup{}
	markup.Inline(
		markup.Row(
			markup.Data("✅ Да", fmt.Sprintf("confirm_hide_%d", wordID)),
			markup.Data("❌ Нет", fmt.Sprintf("cancel_hide_%d", wordID)),
		),
	)

	// Edit message - только edit, никаких send
	if err := c.Edit(text, markup); err != nil {
		h.handleEditError(err, c, userID)
		// Callback уже подтверждён, просто логируем ошибку
	}
	return nil
}

// handleConfirmHideForever permanently hides a word and shows success message with "Ещё" button
func (h *Handler) handleConfirmHideForever(c tele.Context, data string) error {
	userID := c.Sender().ID

	// КРИТИЧЕСКИ ВАЖНО: Отвечаем на callback СРАЗУ
	if c.Callback() != nil {
		if err := c.Respond(); err != nil {
			h.logger.Warn("Failed to acknowledge callback", zap.Error(err))
		}
	}

	// Extract word ID
	data = strings.TrimSpace(data)
	wordIDStr := strings.TrimPrefix(data, "confirm_hide_")
	wordID, err := strconv.Atoi(wordIDStr)
	if err != nil {
		h.logger.Error("Failed to parse word ID", zap.Error(err), zap.String("data", data))
		return nil // Callback уже подтверждён
	}

	// Hide the word forever
	if err := h.wordService.HideWordForever(wordID); err != nil {
		h.logger.Error("Failed to hide word forever", zap.Error(err), zap.Int("word_id", wordID))
		return nil // Callback уже подтверждён
	}

	// Show success message with "Ещё" button
	text := "✅ Слово скрыто навсегда"
	markup := &tele.ReplyMarkup{}
	markup.Inline(markup.Row(btnMore))

	// Edit message - только edit, никаких send
	if err := c.Edit(text, markup); err != nil {
		h.handleEditError(err, c, userID)
		// Callback уже подтверждён, просто логируем ошибку
	}
	return nil
}

// handleCancelHide cancels the hide operation and returns to word display
func (h *Handler) handleCancelHide(c tele.Context, data string) error {
	// Extract word ID for logging (though we don't use it to restore the word)
	data = strings.TrimSpace(data)
	wordIDStr := strings.TrimPrefix(data, "cancel_hide_")
	_, err := strconv.Atoi(wordIDStr)
	if err != nil {
		h.logger.Error("Failed to parse word ID", zap.Error(err), zap.String("data", data))
		return nil // Callback уже подтверждён
	}

	// Show a new random pair (we don't have GetWordByID method to restore the original word)
	return h.handleRandomPair(c)
}

