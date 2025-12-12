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
func (r *WordRepo) GetRandomWord(userID int64) (*domain.Word, error) {
	var w domain.Word
	query := `
		SELECT id, user_id, word, translation, created_at
		FROM words
		WHERE user_id = $1
		ORDER BY RANDOM()
		LIMIT 1
	`
	err := r.db.QueryRow(query, userID).Scan(
		&w.ID, &w.UserID, &w.Word, &w.Translation, &w.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
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
		SELECT id, user_id, word, translation, created_at
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
		if err := rows.Scan(&w.ID, &w.UserID, &w.Word, &w.Translation, &w.CreatedAt); err != nil {
			return nil, err
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

