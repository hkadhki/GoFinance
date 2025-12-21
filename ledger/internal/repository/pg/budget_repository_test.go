package pg

import (
	"context"
	"database/sql"
	"testing"

	"ledger/internal/db/sqlc"
	"ledger/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestBudgetRepo_Upsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	q := sqlc.New(db)
	repo := NewBudgetRepo(q)

	userID := uuid.New()
	budget := domain.Budget{
		Category: "food",
		Limit:    decimal.NewFromInt(300),
		Period:   "monthly",
	}

	mock.ExpectExec(`INSERT INTO budgets`).
		WithArgs(userID, budget.Category, budget.Limit, budget.Period).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Upsert(context.Background(), userID, budget)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBudgetRepo_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	q := sqlc.New(db)
	repo := NewBudgetRepo(q)

	userID := uuid.New()

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "category", "limit_amount", "period",
	}).
		AddRow(1, userID, "food", decimal.NewFromInt(200), "monthly")

	mock.ExpectQuery(`SELECT .* FROM budgets`).
		WithArgs(userID).
		WillReturnRows(rows)

	res, err := repo.List(context.Background(), userID)

	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t, "food", res[0].Category)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBudgetRepo_GetByCategory_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	q := sqlc.New(db)
	repo := NewBudgetRepo(q)

	userID := uuid.New()
	category := "food"

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "category", "limit_amount", "period",
	}).
		AddRow(1, userID, category, decimal.NewFromInt(150), "monthly")

	mock.ExpectQuery(`SELECT .* FROM budgets`).
		WithArgs(userID, category).
		WillReturnRows(rows)

	res, err := repo.GetByCategory(context.Background(), userID, category)

	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, category, res.Category)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBudgetRepo_GetByCategory_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	q := sqlc.New(db)
	repo := NewBudgetRepo(q)

	userID := uuid.New()
	category := "missing"

	mock.ExpectQuery(`SELECT .* FROM budgets`).
		WithArgs(userID, category).
		WillReturnError(sql.ErrNoRows)

	res, err := repo.GetByCategory(context.Background(), userID, category)

	require.NoError(t, err)
	require.Nil(t, res)

	require.NoError(t, mock.ExpectationsWereMet())
}
