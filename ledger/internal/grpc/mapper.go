package grpc

import (
	"context"
	"errors"
	"ledger/internal/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapDomainError(err error) error {
	if err == nil {
		return nil
	}

	var vErr *domain.ValidationError
	if errors.As(err, &vErr) {
		return status.Error(codes.InvalidArgument, vErr.Error())
	}

	var bErr *domain.BudgetExceededError
	if errors.As(err, &bErr) {
		return status.Error(codes.FailedPrecondition, bErr.Error())
	}

	if errors.Is(err, domain.ErrBudgetNotFound) {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return status.Error(codes.DeadlineExceeded, err.Error())
	}

	if errors.Is(err, context.Canceled) {
		return status.Error(codes.Canceled, err.Error())
	}

	return status.Error(codes.Internal, err.Error())
}
