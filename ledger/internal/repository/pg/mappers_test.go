package pg

import (
	"database/sql"
	"testing"
	"time"

	"ledger/internal/db/sqlc"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestMapBudget(t *testing.T) {
	userID := uuid.New()

	b := sqlc.Budget{
		ID:          1,
		UserID:      userID,
		Category:    "food",
		LimitAmount: decimal.NewFromInt(500),
		Period:      "monthly",
	}

	res := mapBudget(b)

	require.Equal(t, b.ID, res.ID)
	require.Equal(t, b.UserID, res.UserID)
	require.Equal(t, b.Category, res.Category)
	require.True(t, b.LimitAmount.Equal(res.Limit))
	require.Equal(t, b.Period, res.Period)
}

func TestMapExpense(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	e := sqlc.Expense{
		ID:       1,
		UserID:   userID,
		Amount:   decimal.NewFromInt(100),
		Category: "food",
		Description: sql.NullString{
			String: "coffee",
			Valid:  true,
		},
		Date: now,
	}

	res := mapExpense(e)

	require.Equal(t, e.ID, res.ID)
	require.Equal(t, e.UserID, res.UserID)
	require.True(t, e.Amount.Equal(res.Amount))
	require.Equal(t, e.Category, res.Category)
	require.Equal(t, "coffee", res.Description)
	require.Equal(t, e.Date, res.Date)
}
