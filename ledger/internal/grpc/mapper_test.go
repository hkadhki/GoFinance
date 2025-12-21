package grpc

import (
	"context"
	"errors"
	"testing"

	"ledger/internal/domain"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMapDomainError_Nil(t *testing.T) {
	err := mapDomainError(nil)
	require.NoError(t, err)
}

func TestMapDomainError_ValidationError(t *testing.T) {
	src := &domain.ValidationError{
		Field:   "amount",
		Message: "must be positive",
	}

	err := mapDomainError(src)

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
	require.Contains(t, st.Message(), "amount")
}

func TestMapDomainError_BudgetExceeded(t *testing.T) {
	src := &domain.BudgetExceededError{
		Category: "food",
	}

	err := mapDomainError(src)

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.FailedPrecondition, st.Code())
	require.Contains(t, st.Message(), "food")
}

func TestMapDomainError_BudgetNotFound(t *testing.T) {
	err := mapDomainError(domain.ErrBudgetNotFound)

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}

func TestMapDomainError_ContextDeadlineExceeded(t *testing.T) {
	err := mapDomainError(context.DeadlineExceeded)

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.DeadlineExceeded, st.Code())
}

func TestMapDomainError_ContextCanceled(t *testing.T) {
	err := mapDomainError(context.Canceled)

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Canceled, st.Code())
}

func TestMapDomainError_UnknownError(t *testing.T) {
	src := errors.New("something went wrong")

	err := mapDomainError(src)

	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Internal, st.Code())
	require.Equal(t, "something went wrong", st.Message())
}
