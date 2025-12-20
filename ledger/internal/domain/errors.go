package domain

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s: %s", e.Field, e.Message)
}

type BudgetExceededError struct {
	Category string
	Limit    decimal.Decimal
	Current  decimal.Decimal
	Amount   decimal.Decimal
}

func (e *BudgetExceededError) Error() string {
	return fmt.Sprintf(
		`budget exceeded for category %s: limit=%s current=%s amount=%s`,
		e.Category,
		e.Limit.StringFixed(2),
		e.Current.StringFixed(2),
		e.Amount.StringFixed(2),
	)
}

type InvalidDateError struct {
	Date string
}

func (e *InvalidDateError) Error() string {
	return fmt.Sprintf("invalid date: %s", e.Date)
}

var ErrBudgetNotFound = errors.New("budget not found")

var ErrUnauthenticated = errors.New("Unauthenticated")
