package pg

import (
	"context"
	"testing"
	"time"

	"ledger/internal/db/sqlc"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestReportRepo_GetReportSummary(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	q := sqlc.New(db)
	repo := NewReportRepo(q)

	userID := uuid.New()
	from := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{"category", "total"}).
		AddRow("food", decimal.NewFromInt(100)).
		AddRow("rent", decimal.NewFromInt(500))

	mock.ExpectQuery(`SELECT .* FROM expenses`).
		WithArgs(userID, from, to).
		WillReturnRows(rows)

	res, err := repo.GetReportSummary(context.Background(), userID, from, to)

	require.NoError(t, err)
	require.Len(t, res, 2)

	require.Equal(t, "food", res[0].Category)
	require.True(t, res[0].Total.Equal(decimal.NewFromInt(100)))

	require.NoError(t, mock.ExpectationsWereMet())
}
