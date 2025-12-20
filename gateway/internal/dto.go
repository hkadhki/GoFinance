package internal

import "github.com/shopspring/decimal"

type CreateTransactionRequest struct {
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Date        string  `json:"date"` // YYYY-MM-DD
}

type TransactionResponse struct {
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
}

type CreateBudgetRequest struct {
	Category string  `json:"category"`
	Limit    float64 `json:"limit"`
	Period   string  `json:"period"`
}

type BudgetResponse struct {
	Category string  `json:"category"`
	Limit    float64 `json:"limit"`
	Period   string  `json:"period"`
}

type ReportResponse struct {
	Category string          `json:"category"`
	Total    decimal.Decimal `json:"total"`
}

type BulkErrorResponse struct {
	Index int    `json:"index"`
	Error string `json:"error"`
}

type BulkImportResponse struct {
	Accepted int64               `json:"accepted"`
	Rejected int64               `json:"rejected"`
	Errors   []BulkErrorResponse `json:"errors"`
}
