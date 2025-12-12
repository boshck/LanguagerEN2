package domain

import "time"

// User represents a bot user
type User struct {
	UserID     int64
	Authorized bool
	CreatedAt  time.Time
}

// UserState represents user's current interaction state
type UserState string

const (
	StateIdle               UserState = "idle"
	StateWaitingWord        UserState = "waiting_word"
	StateWaitingTranslation UserState = "waiting_translation"
	StateWaitingPassword    UserState = "waiting_password"
)

// StateData holds temporary data for user's current state
type StateData struct {
	State       UserState
	CurrentWord string
	MessageID   int // For editing messages
}

