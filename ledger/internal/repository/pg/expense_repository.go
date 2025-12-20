package pg

import (
	"context"
	"database/sql"
	"time"

	"ledger/internal/db/sqlc"
	"ledger/internal/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ExpenseRepo struct {
	q *sqlc.Queries
}

func NewExpenseRepo(q *sqlc.Queries) *ExpenseRepo {
	return &ExpenseRepo{q: q}
}

func (r *ExpenseRepo) Add(
	ctx context.Context,
	userID uuid.UUID,
	t domain.Transaction,
) error {
	_, err := r.q.InsertExpense(ctx, sqlc.InsertExpenseParams{
		UserID:      userID,
		Amount:      t.Amount,
		Category:    t.Category,
		Description: sql.NullString{String: t.Description, Valid: t.Description != ""},
		Date:        t.Date,
	})
	return err
}

func (r *ExpenseRepo) List(
	ctx context.Context,
	userID uuid.UUID,
) ([]domain.Transaction, error) {
	rows, err := r.q.ListExpenses(ctx, userID)
	if err != nil {
		return nil, err
	}

	res := make([]domain.Transaction, 0, len(rows))
	for _, row := range rows {
		res = append(res, mapExpense(row))
	}
	return res, nil
}

func (r *ExpenseRepo) SumByCategory(
	ctx context.Context,
	userID uuid.UUID,
	category string,
) (decimal.Decimal, error) {
	return r.q.GetSumByCategory(ctx, sqlc.GetSumByCategoryParams{
		UserID:   userID,
		Category: category,
	})
}

func (r *ExpenseRepo) SumByCategoryAndPeriod(
	ctx context.Context,
	userID uuid.UUID,
	category string,
	from time.Time,
	to time.Time,
) (decimal.Decimal, error) {
	return r.q.SumByCategoryAndPeriod(ctx, sqlc.SumByCategoryAndPeriodParams{
		UserID:   userID,
		Category: category,
		FromDate: from,
		ToDate:   to,
	})
}
