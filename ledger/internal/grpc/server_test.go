package grpc

import (
	"context"
	"testing"
	"time"

	"ledger/internal/domain"
	ledgerv1 "ledger/ledger/v1"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type mockLedgerService struct {
	addTxFn       func(ctx context.Context, tx domain.Transaction) error
	listTxFn      func(ctx context.Context) ([]domain.Transaction, error)
	setBudgetFn   func(ctx context.Context, b domain.Budget) error
	listBudgetsFn func(ctx context.Context) ([]domain.Budget, error)
	reportFn      func(ctx context.Context, from, to time.Time) ([]domain.ReportSummary, error)
	bulkFn        func(ctx context.Context, txs []domain.Transaction, workers int) (*domain.BulkImportResult, error)
}

func (m *mockLedgerService) AddTransaction(ctx context.Context, tx domain.Transaction) error {
	return m.addTxFn(ctx, tx)
}

func (m *mockLedgerService) ListTransactions(ctx context.Context) ([]domain.Transaction, error) {
	return m.listTxFn(ctx)
}

func (m *mockLedgerService) SetBudget(ctx context.Context, b domain.Budget) error {
	return m.setBudgetFn(ctx, b)
}

func (m *mockLedgerService) ListBudgets(ctx context.Context) ([]domain.Budget, error) {
	return m.listBudgetsFn(ctx)
}

func (m *mockLedgerService) GetReportSummary(ctx context.Context, from, to time.Time) ([]domain.ReportSummary, error) {
	return m.reportFn(ctx, from, to)
}

func (m *mockLedgerService) BulkAddTransactions(
	ctx context.Context,
	txs []domain.Transaction,
	workers int,
) (*domain.BulkImportResult, error) {
	return m.bulkFn(ctx, txs, workers)
}

func TestAddTransaction_OK(t *testing.T) {
	svc := &mockLedgerService{
		addTxFn: func(ctx context.Context, tx domain.Transaction) error {
			require.Equal(t, "food", tx.Category)
			require.True(t, tx.Amount.Equal(decimal.NewFromFloat(100)))
			return nil
		},
	}

	server := NewServer(svc)

	resp, err := server.AddTransaction(context.Background(), &ledgerv1.CreateTransactionRequest{
		Amount:      100,
		Category:    "food",
		Description: "coffee",
		Date:        "2025-01-01",
	})

	require.NoError(t, err)
	require.Equal(t, "food", resp.Category)
	require.Equal(t, 100.0, resp.Amount)
}

func TestAddTransaction_InvalidDate(t *testing.T) {
	server := NewServer(&mockLedgerService{})

	_, err := server.AddTransaction(context.Background(), &ledgerv1.CreateTransactionRequest{
		Amount:   10,
		Category: "food",
		Date:     "invalid-date",
	})

	s, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, s.Code())
}

func TestListTransactions(t *testing.T) {
	svc := &mockLedgerService{
		listTxFn: func(ctx context.Context) ([]domain.Transaction, error) {
			return []domain.Transaction{
				{
					Amount:   decimal.NewFromInt(50),
					Category: "food",
					Date:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			}, nil
		},
	}

	server := NewServer(svc)

	resp, err := server.ListTransactions(context.Background(), &emptypb.Empty{})
	require.NoError(t, err)
	require.Len(t, resp.Transactions, 1)
	require.Equal(t, "food", resp.Transactions[0].Category)
	require.Equal(t, 50.0, resp.Transactions[0].Amount)
}

func TestSetBudget(t *testing.T) {
	svc := &mockLedgerService{
		setBudgetFn: func(ctx context.Context, b domain.Budget) error {
			require.Equal(t, "food", b.Category)
			require.True(t, b.Limit.Equal(decimal.NewFromInt(500)))
			return nil
		},
	}

	server := NewServer(svc)

	resp, err := server.SetBudget(context.Background(), &ledgerv1.CreateBudgetRequest{
		Category: "food",
		Limit:    500,
		Period:   "monthly",
	})

	require.NoError(t, err)
	require.Equal(t, "food", resp.Category)
	require.Equal(t, 500.0, resp.Limit)
}

func TestGetReportSummary(t *testing.T) {
	svc := &mockLedgerService{
		reportFn: func(ctx context.Context, from, to time.Time) ([]domain.ReportSummary, error) {
			return []domain.ReportSummary{
				{
					Category: "food",
					Total:    decimal.NewFromInt(300),
				},
			}, nil
		},
	}

	server := NewServer(svc)

	resp, err := server.GetReportSummary(context.Background(), &ledgerv1.ReportSummaryRequest{
		From: "2025-01-01",
		To:   "2025-01-31",
	})

	require.NoError(t, err)
	require.Equal(t, 300.0, resp.Totals["food"])
}

func TestBulkAddTransactions(t *testing.T) {
	svc := &mockLedgerService{
		bulkFn: func(ctx context.Context, txs []domain.Transaction, workers int) (*domain.BulkImportResult, error) {
			require.Len(t, txs, 2)
			require.Equal(t, 2, workers)

			return &domain.BulkImportResult{
				Accepted: 2,
				Rejected: 0,
			}, nil
		},
	}

	server := NewServer(svc)

	resp, err := server.BulkAddTransactions(context.Background(), &ledgerv1.BulkAddTransactionsRequest{
		Workers: 2,
		Transactions: []*ledgerv1.CreateTransactionRequest{
			{
				Amount:   10,
				Category: "food",
				Date:     "2025-01-01",
			},
			{
				Amount:   20,
				Category: "food",
				Date:     "2025-01-02",
			},
		},
	})

	require.NoError(t, err)
	require.Equal(t, int64(2), resp.Accepted)
	require.Equal(t, int64(0), resp.Rejected)
}
