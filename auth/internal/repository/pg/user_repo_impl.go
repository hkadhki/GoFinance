package pg

import (
	"context"
	"database/sql"

	"auth/internal/repository"

	"github.com/google/uuid"
)

type UserRepo struct {
	db *sql.DB
}

func New(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(
	ctx context.Context,
	email string,
	passwordHash string,
) (string, error) {

	id := uuid.NewString()

	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO users (id, email, password_hash) VALUES ($1,$2,$3)`,
		id,
		email,
		passwordHash,
	)

	return id, err
}

func (r *UserRepo) GetByEmail(
	ctx context.Context,
	email string,
) (*repository.User, error) {

	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, email, password_hash FROM users WHERE email=$1`,
		email,
	)

	var u repository.User
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash); err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *UserRepo) GetByID(
	ctx context.Context,
	id string,
) (*repository.User, error) {

	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, email, password_hash FROM users WHERE id=$1`,
		id,
	)

	var u repository.User
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash); err != nil {
		return nil, err
	}

	return &u, nil
}
