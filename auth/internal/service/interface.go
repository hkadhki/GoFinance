package service

import "context"

type Auth interface {
	Register(ctx context.Context, email, password string) (string, error)
	Login(ctx context.Context, email, password string) (string, error)
	Validate(ctx context.Context, token string) (string, error)
}
