package postgres

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestWordRepo_SaveWord(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewWordRepo(db)

	userID := int64(123)
	word := "hello"
	translation := "привет"

	mock.ExpectExec("INSERT INTO words").
		WithArgs(userID, word, translation).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.SaveWord(userID, word, translation)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWordRepo_GetRandomWord(t *testing.T) {
	tests := []struct {
		name          string
		userID        int64
		mockRows      *sqlmock.Rows
		mockError     error
		expectedNil   bool
		expectedError bool
	}{
		{
			name:   "word found",
			userID: 123,
			mockRows: sqlmock.NewRows([]string{"id", "user_id", "word", "translation", "created_at", "hidden_until", "hidden_forever"}).
				AddRow(1, 123, "hello", "привет", time.Now(), nil, false),
			mockError:     nil,
			expectedNil:   false,
			expectedError: false,
		},
		{
			name:          "no words",
			userID:        456,
			mockRows:      nil,
			mockError:     sql.ErrNoRows,
			expectedNil:   true,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			repo := NewWordRepo(db)

			query := "SELECT id, user_id, word, translation, created_at, hidden_until, hidden_forever FROM words WHERE user_id = \\$1 AND \\(hidden_forever = FALSE OR hidden_forever IS NULL\\) AND \\(hidden_until IS NULL OR hidden_until <= NOW\\(\\)\\)"

			if tt.mockError != nil {
				mock.ExpectQuery(query).WithArgs(tt.userID).WillReturnError(tt.mockError)
			} else {
				mock.ExpectQuery(query).WithArgs(tt.userID).WillReturnRows(tt.mockRows)
			}

			word, err := repo.GetRandomWord(tt.userID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expectedNil {
					assert.Nil(t, word)
				} else {
					assert.NotNil(t, word)
					assert.Equal(t, tt.userID, word.UserID)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestWordRepo_GetDaysWithWords(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewWordRepo(db)

	userID := int64(123)
	limit := 7
	offset := 0

	rows := sqlmock.NewRows([]string{"day", "count"}).
		AddRow(time.Now(), 5).
		AddRow(time.Now().AddDate(0, 0, -1), 3)

	mock.ExpectQuery("SELECT DATE\\(created_at\\)").
		WithArgs(userID, limit, offset).
		WillReturnRows(rows)

	days, err := repo.GetDaysWithWords(userID, limit, offset)

	assert.NoError(t, err)
	assert.Len(t, days, 2)
	assert.Equal(t, 5, days[0].WordCount)
	assert.Equal(t, 3, days[1].WordCount)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWordRepo_GetTotalDaysCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewWordRepo(db)

	userID := int64(123)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(14)

	mock.ExpectQuery("SELECT COUNT\\(DISTINCT DATE\\(created_at\\)\\)").
		WithArgs(userID).
		WillReturnRows(rows)

	count, err := repo.GetTotalDaysCount(userID)

	assert.NoError(t, err)
	assert.Equal(t, 14, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWordRepo_GetWordsByDate(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewWordRepo(db)

	userID := int64(123)
	date := time.Now()

	rows := sqlmock.NewRows([]string{"id", "user_id", "word", "translation", "created_at", "hidden_until", "hidden_forever"}).
		AddRow(1, userID, "hello", "привет", date, nil, false).
		AddRow(2, userID, "world", "мир", date, nil, false)

	mock.ExpectQuery("SELECT id, user_id, word, translation, created_at, hidden_until, hidden_forever FROM words WHERE user_id = \\$1 AND DATE\\(created_at\\) = DATE\\(\\$2\\)").
		WithArgs(userID, date).
		WillReturnRows(rows)

	words, err := repo.GetWordsByDate(userID, date)

	assert.NoError(t, err)
	assert.Len(t, words, 2)
	assert.Equal(t, "hello", words[0].Word)
	assert.Equal(t, "world", words[1].Word)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWordRepo_CleanOldWords(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewWordRepo(db)

	days := 60

	mock.ExpectExec("DELETE FROM words WHERE created_at").
		WithArgs(days).
		WillReturnResult(sqlmock.NewResult(0, 10))

	err = repo.CleanOldWords(days)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWordRepo_HideWordFor7Days(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewWordRepo(db)

	wordID := 1

	mock.ExpectExec("UPDATE words SET hidden_until = NOW\\(\\) \\+ INTERVAL '7 days' WHERE id = \\$1").
		WithArgs(wordID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.HideWordFor7Days(wordID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestWordRepo_HideWordForever(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewWordRepo(db)

	wordID := 1

	mock.ExpectExec("UPDATE words SET hidden_forever = TRUE WHERE id = \\$1").
		WithArgs(wordID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.HideWordForever(wordID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

