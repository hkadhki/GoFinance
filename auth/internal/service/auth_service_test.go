package service

import (
	"auth/internal/jwt"
	"context"
	"errors"
	"testing"

	"auth/internal/domain"
	"auth/internal/repository"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	getByEmail func(ctx context.Context, email string) (*repository.User, error)
	getByID    func(ctx context.Context, id string) (*repository.User, error)
	create     func(ctx context.Context, email, hash string) (string, error)
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*repository.User, error) {
	return m.getByEmail(ctx, email)
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*repository.User, error) {
	return m.getByID(ctx, id)
}

func (m *mockUserRepo) Create(ctx context.Context, email, hash string) (string, error) {
	return m.create(ctx, email, hash)
}

func TestRegister_Success(t *testing.T) {
	repo := &mockUserRepo{
		getByEmail: func(ctx context.Context, email string) (*repository.User, error) {
			return nil, errors.New("not found")
		},
		create: func(ctx context.Context, email, hash string) (string, error) {
			require.NotEmpty(t, hash)
			return "user-123", nil
		},
	}

	svc := New(repo)

	token, err := svc.Register(context.Background(), "test@mail.com", "password")

	require.NoError(t, err)
	require.NotEmpty(t, token)
}

func TestRegister_UserAlreadyExists(t *testing.T) {
	repo := &mockUserRepo{
		getByEmail: func(ctx context.Context, email string) (*repository.User, error) {
			return &repository.User{ID: "1"}, nil
		},
	}

	svc := New(repo)

	_, err := svc.Register(context.Background(), "test@mail.com", "password")

	require.ErrorIs(t, err, ErrUserAlreadyExists)
}

func TestLogin_Success(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte("password"),
		bcrypt.DefaultCost,
	)
	require.NoError(t, err)

	repo := &mockUserRepo{
		getByEmail: func(ctx context.Context, email string) (*repository.User, error) {
			return &repository.User{
				ID:           "user-123",
				PasswordHash: string(hash),
			}, nil
		},
	}

	svc := New(repo)

	token, err := svc.Login(context.Background(), "test@mail.com", "password")

	require.NoError(t, err)
	require.NotEmpty(t, token)
}

func TestLogin_InvalidPassword(t *testing.T) {
	repo := &mockUserRepo{
		getByEmail: func(ctx context.Context, email string) (*repository.User, error) {
			return &repository.User{
				ID:           "user-123",
				PasswordHash: "$2a$10$invalidhash",
			}, nil
		},
	}

	svc := New(repo)

	_, err := svc.Login(context.Background(), "test@mail.com", "wrong")

	require.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestValidate_Success(t *testing.T) {
	repo := &mockUserRepo{
		getByID: func(ctx context.Context, id string) (*repository.User, error) {
			return &repository.User{ID: id}, nil
		},
	}

	svc := New(repo)

	token, err := jwt.Generate("user-123")
	require.NoError(t, err)

	userID, err := svc.Validate(context.Background(), token)

	require.NoError(t, err)
	require.Equal(t, "user-123", userID)
}

func TestValidate_InvalidToken(t *testing.T) {
	repo := &mockUserRepo{}

	svc := New(repo)

	_, err := svc.Validate(context.Background(), "bad.token.value")

	require.ErrorIs(t, err, domain.ErrInvalidToken)
}
