package pg

import (
	"context"
	"database/sql"
	"errors"

	"ledger/internal/db/sqlc"
	"ledger/internal/domain"

	"github.com/google/uuid"
)

type BudgetRepo struct {
	q *sqlc.Queries
}

func NewBudgetRepo(q *sqlc.Queries) *BudgetRepo {
	return &BudgetRepo{q: q}
}

func (r *BudgetRepo) Upsert(
	ctx context.Context,
	userID uuid.UUID,
	b domain.Budget,
) error {
	return r.q.UpsertBudget(ctx, sqlc.UpsertBudgetParams{
		UserID:      userID,
		Category:    b.Category,
		LimitAmount: b.Limit,
		Period:      b.Period,
	})
}

func (r *BudgetRepo) List(
	ctx context.Context,
	userID uuid.UUID,
) ([]domain.Budget, error) {
	rows, err := r.q.ListBudgets(ctx, userID)
	if err != nil {
		return nil, err
	}

	res := make([]domain.Budget, 0, len(rows))
	for _, row := range rows {
		res = append(res, mapBudget(row))
	}
	return res, nil
}

func (r *BudgetRepo) GetByCategory(
	ctx context.Context,
	userID uuid.UUID,
	category string,
) (*domain.Budget, error) {
	row, err := r.q.GetByCategory(ctx, sqlc.GetByCategoryParams{
		UserID:   userID,
		Category: category,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	b := mapBudget(row)
	return &b, nil
}
