package service

import (
	"context"
	"testing"
	"time"

	"ledger/internal/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

type mockBudgetRepo struct {
	budgets map[string]domain.Budget
}

func (m *mockBudgetRepo) Upsert(ctx context.Context, userID uuid.UUID, b domain.Budget) error {
	m.budgets[b.Category] = b
	return nil
}

func (m *mockBudgetRepo) GetByCategory(ctx context.Context, userID uuid.UUID, category string) (*domain.Budget, error) {
	b, ok := m.budgets[category]
	if !ok {
		return nil, nil
	}
	return &b, nil
}

func (m *mockBudgetRepo) List(ctx context.Context, userID uuid.UUID) ([]domain.Budget, error) {
	res := make([]domain.Budget, 0, len(m.budgets))
	for _, b := range m.budgets {
		res = append(res, b)
	}
	return res, nil
}

type mockExpenseRepo struct {
	items []domain.Transaction
}

func (m *mockExpenseRepo) Add(ctx context.Context, userID uuid.UUID, t domain.Transaction) error {
	m.items = append(m.items, t)
	return nil
}

func (m *mockExpenseRepo) List(ctx context.Context, userID uuid.UUID) ([]domain.Transaction, error) {
	return m.items, nil
}

func (m *mockExpenseRepo) BudgetLimit(ctx context.Context, userID uuid.UUID, category string) (decimal.Decimal, error) {
	return decimal.NewFromInt(100), nil
}

func (m *mockExpenseRepo) SumByCategory(ctx context.Context, userID uuid.UUID, category string) (decimal.Decimal, error) {
	sum := decimal.Zero
	for _, t := range m.items {
		if t.Category == category {
			sum = sum.Add(t.Amount)
		}
	}
	return sum, nil
}

func (m *mockExpenseRepo) SumByCategoryAndPeriod(
	ctx context.Context,
	userID uuid.UUID,
	category string,
	from time.Time,
	to time.Time,
) (decimal.Decimal, error) {
	return m.SumByCategory(ctx, userID, category)
}

type mockReportRepo struct{}

func (m *mockReportRepo) GetReportSummary(
	ctx context.Context,
	userID uuid.UUID,
	from time.Time,
	to time.Time,
) ([]domain.ReportSummary, error) {
	return []domain.ReportSummary{
		{
			Category: "food",
			Total:    decimal.NewFromInt(50),
		},
	}, nil
}

func ctxWithUser(userID uuid.UUID) context.Context {
	md := metadata.New(map[string]string{
		"user_id": userID.String(),
	})
	return metadata.NewIncomingContext(context.Background(), md)
}

func TestAddTransaction_OK(t *testing.T) {
	userID := uuid.New()

	budgets := &mockBudgetRepo{
		budgets: map[string]domain.Budget{
			"food": {
				UserID:   userID,
				Category: "food",
				Limit:    decimal.NewFromInt(100),
				Period:   "monthly",
			},
		},
	}

	expenses := &mockExpenseRepo{}
	reports := &mockReportRepo{}

	svc := New(budgets, expenses, reports)

	tx := domain.Transaction{
		Amount:   decimal.NewFromInt(30),
		Category: "food",
		Date:     time.Now(),
	}

	err := svc.AddTransaction(ctxWithUser(userID), tx)
	require.NoError(t, err)

	require.Len(t, expenses.items, 1)
	require.Equal(t, "food", expenses.items[0].Category)
}

func TestAddTransaction_BudgetExceeded(t *testing.T) {
	userID := uuid.New()

	budgets := &mockBudgetRepo{
		budgets: map[string]domain.Budget{
			"food": {
				UserID:   userID,
				Category: "food",
				Limit:    decimal.NewFromInt(10),
				Period:   "monthly",
			},
		},
	}

	expenses := &mockExpenseRepo{}
	reports := &mockReportRepo{}

	svc := New(budgets, expenses, reports)

	tx := domain.Transaction{
		Amount:   decimal.NewFromInt(50),
		Category: "food",
		Date:     time.Now(),
	}

	err := svc.AddTransaction(ctxWithUser(userID), tx)
	require.Error(t, err)
	require.IsType(t, &domain.BudgetExceededError{}, err)
}

func TestListBudgets(t *testing.T) {
	userID := uuid.New()

	budgets := &mockBudgetRepo{
		budgets: map[string]domain.Budget{
			"food": {
				UserID:   userID,
				Category: "food",
				Limit:    decimal.NewFromInt(100),
				Period:   "monthly",
			},
		},
	}

	svc := New(budgets, &mockExpenseRepo{}, &mockReportRepo{})

	res, err := svc.ListBudgets(ctxWithUser(userID))
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t, "food", res[0].Category)
}

func TestGetReportSummary(t *testing.T) {
	userID := uuid.New()

	budgets := &mockBudgetRepo{
		budgets: map[string]domain.Budget{
			"food": {
				UserID:   userID,
				Category: "food",
				Limit:    decimal.NewFromInt(100),
				Period:   "monthly",
			},
		},
	}

	expenses := &mockExpenseRepo{
		items: []domain.Transaction{
			{
				UserID:   userID,
				Category: "food",
				Amount:   decimal.NewFromInt(50),
				Date:     time.Now(),
			},
		},
	}

	svc := New(budgets, expenses, &mockReportRepo{})

	from := time.Now().AddDate(0, 0, -7)
	to := time.Now()

	res, err := svc.GetReportSummary(ctxWithUser(userID), from, to)

	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t, "food", res[0].Category)
	require.True(t, res[0].Total.Equal(decimal.NewFromInt(50)))
}

func TestUserIDFromContext(t *testing.T) {
	id := uuid.New()
	ctx := ctxWithUser(id)

	got, err := UserIDFromContext(ctx)
	require.NoError(t, err)
	require.Equal(t, id, got)
}
