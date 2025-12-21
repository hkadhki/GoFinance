package grpc

import (
	"auth/internal/domain"
	"context"

	authv1 "auth/auth/v1"
	"auth/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	authv1.UnimplementedAuthServiceServer
	auth service.Auth
}

func New(authService service.Auth) *Server {
	return &Server{
		auth: authService,
	}
}

func (s *Server) Register(
	ctx context.Context,
	req *authv1.RegisterRequest,
) (*authv1.AuthResponse, error) {

	token, err := s.auth.Register(ctx, req.Email, req.Password)
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.AuthResponse{
		AccessToken: token,
	}, nil
}

func (s *Server) Login(
	ctx context.Context,
	req *authv1.LoginRequest,
) (*authv1.AuthResponse, error) {

	token, err := s.auth.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.AuthResponse{
		AccessToken: token,
	}, nil
}

func (s *Server) Validate(
	ctx context.Context,
	req *authv1.ValidateRequest,
) (*authv1.ValidateResponse, error) {

	userID, err := s.auth.Validate(ctx, req.Token)
	if err != nil {
		return &authv1.ValidateResponse{
			Valid: false,
		}, nil
	}

	return &authv1.ValidateResponse{
		UserId: userID,
		Valid:  true,
	}, nil
}

func mapError(err error) error {
	switch err {
	case service.ErrInvalidCredentials:
		return status.Error(codes.Unauthenticated, "invalid credentials")
	case service.ErrUserAlreadyExists:
		return status.Error(codes.AlreadyExists, "user already exists")
	case domain.ErrInvalidToken:
		return status.Error(codes.Unauthenticated, "invalid token")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
