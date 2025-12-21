package jwt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateAndValidate(t *testing.T) {
	token, err := Generate("user-123")
	require.NoError(t, err)

	userID, valid, err := Validate(token)
	require.NoError(t, err)
	require.True(t, valid)
	require.Equal(t, "user-123", userID)
}

func TestValidate_InvalidToken(t *testing.T) {
	_, valid, err := Validate("invalid.token")

	require.Error(t, err)
	require.False(t, valid)
}
