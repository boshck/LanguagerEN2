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

	// Callback processing locks per user (prevents race conditions)
	callbackLocks map[int64]*sync.Mutex
	callbackMux   sync.RWMutex
}

// NewHandler creates a new handler instance
func NewHandler(
	bot *tele.Bot,
	authService *service.AuthService,
	wordService *service.WordService,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		bot:           bot,
		authService:   authService,
		wordService:   wordService,
		logger:        logger,
		states:        make(map[int64]*domain.StateData),
		callbackLocks: make(map[int64]*sync.Mutex),
	}
}

// RegisterHandlers registers all bot handlers
func (h *Handler) RegisterHandlers() {
	// Add middleware to log ALL updates
	h.bot.Use(func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			callback := c.Callback()
			if callback != nil {
				h.logger.Info("UPDATE: CallbackQuery received",
					zap.String("data", callback.Data),
					zap.String("id", callback.ID),
					zap.String("unique", callback.Unique),
					zap.Int64("user_id", c.Sender().ID),
				)
			} else if c.Message() != nil {
				h.logger.Info("UPDATE: Message received",
					zap.String("text", c.Message().Text),
					zap.Int64("user_id", c.Sender().ID),
				)
			} else {
				h.logger.Info("UPDATE: Other update type",
					zap.Int64("user_id", c.Sender().ID),
				)
			}
			return next(c)
		}
	})

	// Commands
	h.bot.Handle("/start", h.handleStart)

	// Text messages
	h.bot.Handle(tele.OnText, h.handleText)

	// Generic callback handler for ALL callbacks
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

