package domain

import (
	"strings"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Budget struct {
	ID       int32           `json:"id"`
	UserID   uuid.UUID       `json:"user_id"`
	Category string          `json:"category"`
	Limit    decimal.Decimal `json:"limit"`
	Period   string          `json:"period"` // daily | weekly | monthly | ""
}

func (b Budget) Validate() error {
	if strings.TrimSpace(b.Category) == "" {
		return &ValidationError{
			Field:   "category",
			Message: "must not be empty",
		}
	}
	if b.Limit.LessThanOrEqual(decimal.Zero) {
		return &ValidationError{
			Field:   "limit",
			Message: "must be positive",
		}
	}
	if b.Period != "" && b.Period != "monthly" && b.Period != "weekly" && b.Period != "daily" {
		return &ValidationError{
			Field:   "period",
			Message: "can be either daily , monthly or weekly",
		}
	}
	return nil
}
