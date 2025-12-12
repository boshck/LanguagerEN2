package postgres

import (
	"database/sql"
	"time"

	"languager/internal/domain"
)

// WordRepo implements repository.WordRepository
type WordRepo struct {
	db *sql.DB
}

// NewWordRepo creates a new word repository
func NewWordRepo(db *sql.DB) *WordRepo {
	return &WordRepo{db: db}
}

// SaveWord saves a word-translation pair
func (r *WordRepo) SaveWord(userID int64, word, translation string) error {
	query := `
		INSERT INTO words (user_id, word, translation)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.Exec(query, userID, word, translation)
	return err
}

// GetRandomWord returns a random word for the user
// Excludes words that are hidden forever or hidden until a future date
func (r *WordRepo) GetRandomWord(userID int64) (*domain.Word, error) {
	var w domain.Word
	var hiddenUntil sql.NullTime
	query := `
		SELECT id, user_id, word, translation, created_at, hidden_until, hidden_forever
		FROM words
		WHERE user_id = $1
			AND (hidden_forever = FALSE OR hidden_forever IS NULL)
			AND (hidden_until IS NULL OR hidden_until <= NOW())
		ORDER BY RANDOM()
		LIMIT 1
	`
	err := r.db.QueryRow(query, userID).Scan(
		&w.ID, &w.UserID, &w.Word, &w.Translation, &w.CreatedAt, &hiddenUntil, &w.HiddenForever,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if hiddenUntil.Valid {
		w.HiddenUntil = &hiddenUntil.Time
	}

	return &w, nil
}

// GetDaysWithWords returns days that have words with counts
func (r *WordRepo) GetDaysWithWords(userID int64, limit, offset int) ([]domain.Day, error) {
	query := `
		SELECT DATE(created_at) as day, COUNT(*) as count
		FROM words
		WHERE user_id = $1 AND created_at >= NOW() - INTERVAL '60 days'
		GROUP BY DATE(created_at)
		ORDER BY day DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var days []domain.Day
	for rows.Next() {
		var d domain.Day
		if err := rows.Scan(&d.Date, &d.WordCount); err != nil {
			return nil, err
		}
		days = append(days, d)
	}

	return days, rows.Err()
}

// GetTotalDaysCount returns total number of days with words
func (r *WordRepo) GetTotalDaysCount(userID int64) (int, error) {
	query := `
		SELECT COUNT(DISTINCT DATE(created_at))
		FROM words
		WHERE user_id = $1 AND created_at >= NOW() - INTERVAL '60 days'
	`

	var count int
	err := r.db.QueryRow(query, userID).Scan(&count)
	return count, err
}

// GetWordsByDate returns all words for a specific date
func (r *WordRepo) GetWordsByDate(userID int64, date time.Time) ([]domain.Word, error) {
	query := `
		SELECT id, user_id, word, translation, created_at, hidden_until, hidden_forever
		FROM words
		WHERE user_id = $1 AND DATE(created_at) = DATE($2)
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var words []domain.Word
	for rows.Next() {
		var w domain.Word
		var hiddenUntil sql.NullTime
		if err := rows.Scan(&w.ID, &w.UserID, &w.Word, &w.Translation, &w.CreatedAt, &hiddenUntil, &w.HiddenForever); err != nil {
			return nil, err
		}
		if hiddenUntil.Valid {
			w.HiddenUntil = &hiddenUntil.Time
		}
		words = append(words, w)
	}

	return words, rows.Err()
}

// CleanOldWords deletes words older than specified days
func (r *WordRepo) CleanOldWords(days int) error {
	query := `
		DELETE FROM words
		WHERE created_at < NOW() - INTERVAL '1 day' * $1
	`
	_, err := r.db.Exec(query, days)
	return err
}

// HideWordFor7Days hides a word from random pair for 7 days
func (r *WordRepo) HideWordFor7Days(wordID int) error {
	query := `
		UPDATE words
		SET hidden_until = NOW() + INTERVAL '7 days'
		WHERE id = $1
	`
	_, err := r.db.Exec(query, wordID)
	return err
}

// HideWordForever permanently hides a word from random pair
func (r *WordRepo) HideWordForever(wordID int) error {
	query := `
		UPDATE words
		SET hidden_forever = TRUE
		WHERE id = $1
	`
	_, err := r.db.Exec(query, wordID)
	return err
}

