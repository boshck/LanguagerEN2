package postgres

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestUserRepo_IsAuthorized(t *testing.T) {
	tests := []struct {
		name          string
		userID        int64
		mockRows      *sqlmock.Rows
		mockError     error
		expectedAuth  bool
		expectedError bool
	}{
		{
			name:          "authorized user",
			userID:        123,
			mockRows:      sqlmock.NewRows([]string{"authorized"}).AddRow(true),
			mockError:     nil,
			expectedAuth:  true,
			expectedError: false,
		},
		{
			name:          "unauthorized user",
			userID:        456,
			mockRows:      sqlmock.NewRows([]string{"authorized"}).AddRow(false),
			mockError:     nil,
			expectedAuth:  false,
			expectedError: false,
		},
		{
			name:          "user not exists",
			userID:        789,
			mockRows:      nil,
			mockError:     sql.ErrNoRows,
			expectedAuth:  false,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			repo := NewUserRepo(db)

			query := "SELECT authorized FROM users WHERE user_id = \\$1"

			if tt.mockError != nil {
				mock.ExpectQuery(query).WithArgs(tt.userID).WillReturnError(tt.mockError)
			} else {
				mock.ExpectQuery(query).WithArgs(tt.userID).WillReturnRows(tt.mockRows)
			}

			authorized, err := repo.IsAuthorized(tt.userID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedAuth, authorized)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepo_AuthorizeUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepo(db)

	userID := int64(123)

	// Only userID is a parameter, TRUE is a SQL constant
	mock.ExpectExec("INSERT INTO users").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.AuthorizeUser(userID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepo_EnsureUserExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserRepo(db)

	userID := int64(123)

	// Only userID is a parameter, FALSE is a SQL constant
	mock.ExpectExec("INSERT INTO users").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.EnsureUserExists(userID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

