package service

import (
	"auth/internal/domain"
	"context"
	"errors"

	"auth/internal/jwt"
	"auth/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
)

type AuthService struct {
	users repository.UserRepository
}

func New(users repository.UserRepository) *AuthService {
	return &AuthService{
		users: users,
	}
}

func (s *AuthService) Register(
	ctx context.Context,
	email string,
	password string,
) (string, error) {

	_, err := s.users.GetByEmail(ctx, email)
	if err == nil {
		return "", ErrUserAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", err
	}

	userID, err := s.users.Create(ctx, email, string(hash))
	if err != nil {
		return "", err
	}

	return jwt.Generate(userID)
}

func (s *AuthService) Login(
	ctx context.Context,
	email string,
	password string,
) (string, error) {

	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	); err != nil {
		return "", ErrInvalidCredentials
	}

	return jwt.Generate(user.ID)
}
func (s *AuthService) Validate(ctx context.Context, token string) (string, error) {
	userID, valid, err := jwt.Validate(token)
	if err != nil || !valid {
		return "", domain.ErrInvalidToken
	}

	_, err = s.users.GetByID(ctx, userID)
	if err != nil {
		return "", domain.ErrInvalidToken
	}

	return userID, nil
}
