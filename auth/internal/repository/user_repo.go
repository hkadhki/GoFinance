package repository

import "context"

type User struct {
	ID           string
	Email        string
	PasswordHash string
}

type UserRepository interface {
	Create(ctx context.Context, email, passwordHash string) (string, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
}
