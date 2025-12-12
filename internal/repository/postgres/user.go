package postgres

import (
	"database/sql"
)

// UserRepo implements repository.UserRepository
type UserRepo struct {
	db *sql.DB
}

// NewUserRepo creates a new user repository
func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

// IsAuthorized checks if user is authorized
func (r *UserRepo) IsAuthorized(userID int64) (bool, error) {
	var authorized bool
	query := `SELECT authorized FROM users WHERE user_id = $1`
	err := r.db.QueryRow(query, userID).Scan(&authorized)

	if err == sql.ErrNoRows {
		// User doesn't exist yet
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return authorized, nil
}

// AuthorizeUser marks user as authorized
func (r *UserRepo) AuthorizeUser(userID int64) error {
	query := `
		INSERT INTO users (user_id, authorized)
		VALUES ($1, TRUE)
		ON CONFLICT (user_id)
		DO UPDATE SET authorized = TRUE
	`
	_, err := r.db.Exec(query, userID)
	return err
}

// EnsureUserExists creates user if not exists
func (r *UserRepo) EnsureUserExists(userID int64) error {
	query := `
		INSERT INTO users (user_id, authorized)
		VALUES ($1, FALSE)
		ON CONFLICT (user_id) DO NOTHING
	`
	_, err := r.db.Exec(query, userID)
	return err
}

