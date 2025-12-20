package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Transaction struct {
	ID          int32           `json:"id"`
	UserID      uuid.UUID       `json:"user_id"`
	Amount      decimal.Decimal `json:"amount"`
	Category    string          `json:"category"`
	Description string          `json:"description"`
	Date        time.Time       `json:"date"`
}

func (t Transaction) Validate() error {
	if t.Amount.LessThanOrEqual(decimal.Zero) {
		return &ValidationError{
			Field:   "amount",
			Message: "must be positive",
		}
	}

	if strings.TrimSpace(t.Category) == "" {
		return &ValidationError{
			Field:   "category",
			Message: "must not be empty",
		}
	}

	if t.Date.IsZero() {
		return &ValidationError{
			Field:   "date",
			Message: "must not be empty",
		}
	}
	return nil
}
