package domain

import "time"

// Word represents a word-translation pair
type Word struct {
	ID           int
	UserID       int64
	Word         string
	Translation  string
	CreatedAt    time.Time
	HiddenUntil  *time.Time
	HiddenForever bool
}

// WordPair is a simplified version for display
type WordPair struct {
	Word        string
	Translation string
}

