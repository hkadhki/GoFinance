package grpc

import (
	"context"
	"errors"
	"testing"

	authv1 "auth/auth/v1"
	"auth/internal/service"

	"github.com/stretchr/testify/require"
)

type mockAuthService struct {
	register func(ctx context.Context, email, password string) (string, error)
	login    func(ctx context.Context, email, password string) (string, error)
	validate func(ctx context.Context, token string) (string, error)
}

func (m *mockAuthService) Register(ctx context.Context, email, password string) (string, error) {
	return m.register(ctx, email, password)
}

func (m *mockAuthService) Login(ctx context.Context, email, password string) (string, error) {
	return m.login(ctx, email, password)
}

func (m *mockAuthService) Validate(ctx context.Context, token string) (string, error) {
	return m.validate(ctx, token)
}

func TestRegister_Success(t *testing.T) {
	svc := &mockAuthService{
		register: func(ctx context.Context, email, password string) (string, error) {
			return "jwt-token", nil
		},
	}

	server := New((*service.AuthService)(nil))
	server.auth = svc

	resp, err := server.Register(context.Background(), &authv1.RegisterRequest{
		Email:    "test@mail.com",
		Password: "password",
	})

	require.NoError(t, err)
	require.Equal(t, "jwt-token", resp.AccessToken)
}

func TestRegister_UserExists(t *testing.T) {
	svc := &mockAuthService{
		register: func(ctx context.Context, email, password string) (string, error) {
			return "", service.ErrUserAlreadyExists
		},
	}

	server := New((*service.AuthService)(nil))
	server.auth = svc

	_, err := server.Register(context.Background(), &authv1.RegisterRequest{
		Email:    "test@mail.com",
		Password: "password",
	})

	require.Error(t, err)
}

func TestLogin_Success(t *testing.T) {
	svc := &mockAuthService{
		login: func(ctx context.Context, email, password string) (string, error) {
			return "jwt-token", nil
		},
	}

	server := New((*service.AuthService)(nil))
	server.auth = svc

	resp, err := server.Login(context.Background(), &authv1.LoginRequest{
		Email:    "test@mail.com",
		Password: "password",
	})

	require.NoError(t, err)
	require.Equal(t, "jwt-token", resp.AccessToken)
}

func TestLogin_InvalidCredentials(t *testing.T) {
	svc := &mockAuthService{
		login: func(ctx context.Context, email, password string) (string, error) {
			return "", service.ErrInvalidCredentials
		},
	}

	server := New((*service.AuthService)(nil))
	server.auth = svc

	_, err := server.Login(context.Background(), &authv1.LoginRequest{
		Email:    "test@mail.com",
		Password: "bad",
	})

	require.Error(t, err)
}

func TestValidate_Success(t *testing.T) {
	svc := &mockAuthService{
		validate: func(ctx context.Context, token string) (string, error) {
			return "user-123", nil
		},
	}

	server := New((*service.AuthService)(nil))
	server.auth = svc

	resp, err := server.Validate(context.Background(), &authv1.ValidateRequest{
		Token: "jwt-token",
	})

	require.NoError(t, err)
	require.True(t, resp.Valid)
	require.Equal(t, "user-123", resp.UserId)
}

func TestValidate_Invalid(t *testing.T) {
	svc := &mockAuthService{
		validate: func(ctx context.Context, token string) (string, error) {
			return "", errors.New("invalid token")
		},
	}

	server := New((*service.AuthService)(nil))
	server.auth = svc

	resp, err := server.Validate(context.Background(), &authv1.ValidateRequest{
		Token: "bad-token",
	})

	require.NoError(t, err)
	require.False(t, resp.Valid)
}
