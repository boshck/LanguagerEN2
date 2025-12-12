package handler

import (
	"fmt"
	"strconv"
	"strings"
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

	// Get first page
	days, totalPages, err := h.wordService.GetDaysList(userID, 1)
	if err != nil {
		h.logger.Error("Failed to get days list", zap.Error(err))
		return c.Respond(&tele.CallbackResponse{Text: "Ошибка при загрузке данных"})
	}

	if len(days) == 0 {
		return c.Respond(&tele.CallbackResponse{
			Text:      "У тебя пока нет сохранённых слов",
			ShowAlert: true,
		})
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

	// Edit message if callback, send new if command
	if c.Callback() != nil {
		if err := c.Edit(text, markup); err != nil {
			// If can't edit (message too old), send new
			return c.Send(text, markup)
		}
		return c.Respond()
	}
	return c.Send(text, markup)
}

// handleRandomPair shows a random word-translation pair
func (h *Handler) handleRandomPair(c tele.Context) error {
	userID := c.Sender().ID

	word, err := h.wordService.GetRandomPair(userID)
	if err != nil {
		h.logger.Error("Failed to get random word", zap.Error(err))
		return c.Respond(&tele.CallbackResponse{Text: "Ошибка при загрузке"})
	}

	if word == nil {
		return c.Respond(&tele.CallbackResponse{
			Text:      "У тебя пока нет сохранённых слов",
			ShowAlert: true,
		})
	}

	text := fmt.Sprintf("🎲 Случайная пара:\n\n📝 %s\n🔄 %s", word.Word, word.Translation)

	markup := &tele.ReplyMarkup{}
	markup.Inline(
		markup.Row(btnMore),
		markup.Row(btnBack),
	)

	// Edit message if callback, send new if command
	if c.Callback() != nil {
		if err := c.Edit(text, markup); err != nil {
			// If can't edit (message too old), send new
			return c.Send(text, markup)
		}
		return c.Respond()
	}
	return c.Send(text, markup)
}

// handleCancel cancels current operation and resets state
func (h *Handler) handleCancel(c tele.Context) error {
	userID := c.Sender().ID

	h.ResetState(userID)

	if err := c.Edit(
		"🏠 Главное меню\n\nВыберите действие:",
		mainMenuMarkup(),
	); err != nil {
		return err
	}
	return c.Respond()
}

// handlePagination handles page navigation
func (h *Handler) handlePagination(c tele.Context, data string) error {
	userID := c.Sender().ID

	// Extract page number - trim whitespace first
	data = strings.TrimSpace(data)
	pageStr := strings.TrimPrefix(data, "page_")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return c.Respond(&tele.CallbackResponse{Text: "Неверная страница"})
	}

	days, totalPages, err := h.wordService.GetDaysList(userID, page)
	if err != nil {
		h.logger.Error("Failed to get days list", zap.Error(err))
		return c.Respond(&tele.CallbackResponse{Text: "Ошибка при загрузке"})
	}

	if len(days) == 0 {
		return c.Respond(&tele.CallbackResponse{Text: "Нет данных"})
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

	if err := c.Edit(text, markup); err != nil {
		// If can't edit (message too old), acknowledge callback first, then send new
		if ackErr := c.Respond(); ackErr != nil {
			h.logger.Warn("Failed to acknowledge callback", zap.Error(ackErr))
		}
		return c.Send(text, markup)
	}
	return c.Respond()
}

// handleDaySelection shows words for selected day
func (h *Handler) handleDaySelection(c tele.Context, data string) error {
	userID := c.Sender().ID

	// Extract date - trim whitespace first, then remove prefix
	data = strings.TrimSpace(data)
	dateStr := strings.TrimPrefix(data, "day_")
	h.logger.Info("Handling day selection", zap.String("date", dateStr), zap.String("original_data", data), zap.Int64("user_id", userID))

	words, err := h.wordService.GetWordsByDate(userID, dateStr)
	if err != nil {
		h.logger.Error("Failed to get words by date", zap.Error(err))
		return c.Respond(&tele.CallbackResponse{Text: "Ошибка при загрузке"})
	}

	if len(words) == 0 {
		return c.Respond(&tele.CallbackResponse{Text: "Нет слов за этот день"})
	}

	// Build message with all words
	text := fmt.Sprintf("📝 Слова за выбранный день (%d):\n\n", len(words))
	for i, word := range words {
		text += fmt.Sprintf("%d. %s — %s\n\n", i+1, word.Word, word.Translation)
	}

	markup := &tele.ReplyMarkup{}
	markup.Inline(
		markup.Row(btnBackToDays, btnMainMenu),
	)

	if err := c.Edit(text, markup); err != nil {
		return err
	}
	return c.Respond()
}

