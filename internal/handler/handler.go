package handler

import (
	"sync"

	"languager/internal/domain"
	"languager/internal/service"

	"go.uber.org/zap"
	tele "gopkg.in/telebot.v3"
)

// Handler manages all bot interactions
type Handler struct {
	bot         *tele.Bot
	authService *service.AuthService
	wordService *service.WordService
	logger      *zap.Logger

	// User states (in-memory state machine)
	states   map[int64]*domain.StateData
	stateMux sync.RWMutex
}

// NewHandler creates a new handler instance
func NewHandler(
	bot *tele.Bot,
	authService *service.AuthService,
	wordService *service.WordService,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		bot:         bot,
		authService: authService,
		wordService: wordService,
		logger:      logger,
		states:      make(map[int64]*domain.StateData),
	}
}

// RegisterHandlers registers all bot handlers
func (h *Handler) RegisterHandlers() {
	// Commands
	h.bot.Handle("/start", h.handleStart)

	// Text messages
	h.bot.Handle(tele.OnText, h.handleText)

	// Callback queries (inline buttons)
	h.bot.Handle(&btnViewDays, h.handleViewDays)
	h.bot.Handle(&btnRandomPair, h.handleRandomPair)
	h.bot.Handle(&btnCancel, h.handleCancel)
	h.bot.Handle(&btnMore, h.handleRandomPair)
	h.bot.Handle(&btnBack, h.handleStart)
	h.bot.Handle(&btnBackToDays, h.handleViewDays)
	h.bot.Handle(&btnMainMenu, h.handleStart)

	// Generic callback handler for dynamic data
	h.bot.Handle(tele.OnCallback, h.handleCallback)
}

// GetState returns user's current state
func (h *Handler) GetState(userID int64) *domain.StateData {
	h.stateMux.RLock()
	defer h.stateMux.RUnlock()

	state, exists := h.states[userID]
	if !exists {
		return &domain.StateData{State: domain.StateIdle}
	}
	return state
}

// SetState sets user's state
func (h *Handler) SetState(userID int64, state *domain.StateData) {
	h.stateMux.Lock()
	defer h.stateMux.Unlock()
	h.states[userID] = state
}

// ResetState resets user to idle state
func (h *Handler) ResetState(userID int64) {
	h.SetState(userID, &domain.StateData{State: domain.StateIdle})
}

// Inline keyboard buttons
var (
	btnViewDays = tele.Btn{
		Unique: "view_days",
		Text:   "📅 Посмотреть дни",
	}
	btnRandomPair = tele.Btn{
		Unique: "random_pair",
		Text:   "🎲 Случайная пара",
	}
	btnCancel = tele.Btn{
		Unique: "cancel",
		Text:   "❌ Отменить",
	}
	btnMore = tele.Btn{
		Unique: "more",
		Text:   "🔄 Ещё",
	}
	btnBack = tele.Btn{
		Unique: "back",
		Text:   "🏠 Назад",
	}
	btnBackToDays = tele.Btn{
		Unique: "back_to_days",
		Text:   "◀️ К дням",
	}
	btnMainMenu = tele.Btn{
		Unique: "main_menu",
		Text:   "🏠 Главное меню",
	}
)

// mainMenuMarkup returns the main menu keyboard
func mainMenuMarkup() *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}
	menu.Inline(
		menu.Row(btnViewDays),
		menu.Row(btnRandomPair),
	)
	return menu
}

