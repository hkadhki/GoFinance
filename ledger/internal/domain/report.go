package domain

import "github.com/shopspring/decimal"

type ReportSummary struct {
	Category string          `json:"category"`
	Total    decimal.Decimal `json:"total"`
}
