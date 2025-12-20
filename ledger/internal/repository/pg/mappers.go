package pg

import (
	"ledger/internal/db/sqlc"
	"ledger/internal/domain"
)

func mapBudget(b sqlc.Budget) domain.Budget {
	return domain.Budget{
		ID:       b.ID,
		UserID:   b.UserID,
		Category: b.Category,
		Limit:    b.LimitAmount,
		Period:   b.Period,
	}
}

func mapExpense(e sqlc.Expense) domain.Transaction {
	return domain.Transaction{
		ID:          e.ID,
		UserID:      e.UserID,
		Amount:      e.Amount,
		Category:    e.Category,
		Description: e.Description.String,
		Date:        e.Date,
	}
}
