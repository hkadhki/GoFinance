package pg

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"ledger/internal/db/sqlc"
	"ledger/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestExpenseRepo_Add(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	q := sqlc.New(db)
	repo := NewExpenseRepo(q)

	userID := uuid.New()
	tx := domain.Transaction{
		Amount:      decimal.NewFromInt(100),
		Category:    "food",
		Description: "lunch",
		Date:        time.Now(),
	}

	mock.ExpectQuery(`INSERT INTO expenses`).
		WithArgs(
			userID,
			tx.Amount,
			tx.Category,
			sql.NullString{String: tx.Description, Valid: true},
			tx.Date,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err = repo.Add(context.Background(), userID, tx)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExpenseRepo_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	q := sqlc.New(db)
	repo := NewExpenseRepo(q)

	userID := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "amount", "category", "description", "date",
	}).AddRow(
		1, userID, decimal.NewFromInt(50), "food", "pizza", now,
	)

	mock.ExpectQuery(`SELECT .* FROM expenses`).
		WithArgs(userID).
		WillReturnRows(rows)

	res, err := repo.List(context.Background(), userID)
	require.NoError(t, err)
	require.Len(t, res, 1)

	require.Equal(t, "food", res[0].Category)
	require.Equal(t, "pizza", res[0].Description)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExpenseRepo_SumByCategory(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	q := sqlc.New(db)
	repo := NewExpenseRepo(q)

	userID := uuid.New()
	category := "food"

	mock.ExpectQuery(`SELECT COALESCE\(SUM\(amount\)`).
		WithArgs(userID, category).
		WillReturnRows(
			sqlmock.NewRows([]string{"sum"}).
				AddRow(decimal.NewFromInt(150)),
		)

	sum, err := repo.SumByCategory(context.Background(), userID, category)
	require.NoError(t, err)
	require.True(t, sum.Equal(decimal.NewFromInt(150)))

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExpenseRepo_SumByCategoryAndPeriod(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	q := sqlc.New(db)
	repo := NewExpenseRepo(q)

	userID := uuid.New()
	category := "food"
	from := time.Now().AddDate(0, 0, -7)
	to := time.Now()

	mock.ExpectQuery(`SELECT COALESCE\(SUM\(amount\)`).
		WithArgs(userID, category, from, to).
		WillReturnRows(
			sqlmock.NewRows([]string{"sum"}).
				AddRow(decimal.NewFromInt(200)),
		)

	sum, err := repo.SumByCategoryAndPeriod(
		context.Background(),
		userID,
		category,
		from,
		to,
	)

	require.NoError(t, err)
	require.True(t, sum.Equal(decimal.NewFromInt(200)))

	require.NoError(t, mock.ExpectationsWereMet())
}
