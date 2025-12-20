package service

import (
	"context"
	domain2 "ledger/internal/domain"
	"time"
)

type LedgerService interface {
	AddTransaction(ctx context.Context, t domain2.Transaction) error
	ListTransactions(ctx context.Context) ([]domain2.Transaction, error)
	SetBudget(ctx context.Context, b domain2.Budget) error
	ListBudgets(ctx context.Context) ([]domain2.Budget, error)
	GetReportSummary(ctx context.Context, from time.Time, to time.Time) ([]domain2.ReportSummary, error)
	BulkAddTransactions(ctx context.Context, txs []domain2.Transaction, workers int) (*domain2.BulkImportResult, error)
}
