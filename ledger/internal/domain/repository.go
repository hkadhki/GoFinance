package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type BudgetRepository interface {
	Upsert(
		ctx context.Context,
		userID uuid.UUID,
		b Budget,
	) error

	GetByCategory(
		ctx context.Context,
		userID uuid.UUID,
		category string,
	) (*Budget, error)

	List(
		ctx context.Context,
		userID uuid.UUID,
	) ([]Budget, error)
}

type ExpenseRepository interface {
	Add(
		ctx context.Context,
		userID uuid.UUID,
		t Transaction,
	) error

	List(
		ctx context.Context,
		userID uuid.UUID,
	) ([]Transaction, error)

	SumByCategory(
		ctx context.Context,
		userID uuid.UUID,
		category string,
	) (decimal.Decimal, error)

	SumByCategoryAndPeriod(
		ctx context.Context,
		userID uuid.UUID,
		category string,
		from time.Time,
		to time.Time,
	) (decimal.Decimal, error)
}

type ReportRepository interface {
	GetReportSummary(
		ctx context.Context,
		userID uuid.UUID,
		from time.Time,
		to time.Time,
	) ([]ReportSummary, error)
}
