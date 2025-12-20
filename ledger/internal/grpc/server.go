package grpc

import (
	"context"
	domain2 "ledger/internal/domain"
	"ledger/internal/service"
	ledgerv1 "ledger/ledger/v1"
	"time"

	"github.com/shopspring/decimal"
	_ "github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	ledgerv1.UnimplementedLedgerServiceServer
	service service.LedgerService
}

func NewServer(s service.LedgerService) *Server {
	return &Server{service: s}
}

func (s *Server) AddTransaction(
	ctx context.Context,
	req *ledgerv1.CreateTransactionRequest,
) (*ledgerv1.Transaction, error) {

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid date")
	}

	tx := domain2.Transaction{
		Amount:      decimal.NewFromFloat(req.Amount),
		Category:    req.Category,
		Description: req.Description,
		Date:        date,
	}

	if err := s.service.AddTransaction(ctx, tx); err != nil {
		return nil, mapDomainError(err)
	}

	return &ledgerv1.Transaction{
		Amount:      req.Amount,
		Category:    tx.Category,
		Description: tx.Description,
		Date:        req.Date,
	}, nil
}

func (s *Server) ListTransactions(
	ctx context.Context,
	_ *emptypb.Empty,
) (*ledgerv1.ListTransactionsResponse, error) {

	txs, err := s.service.ListTransactions(ctx)
	if err != nil {
		return nil, mapDomainError(err)
	}

	res := &ledgerv1.ListTransactionsResponse{}
	for _, t := range txs {
		res.Transactions = append(res.Transactions, &ledgerv1.Transaction{
			Amount:      t.Amount.InexactFloat64(),
			Category:    t.Category,
			Description: t.Description,
			Date:        t.Date.Format("2006-01-02"),
		})
	}

	return res, nil
}

func (s *Server) SetBudget(
	ctx context.Context,
	req *ledgerv1.CreateBudgetRequest,
) (*ledgerv1.Budget, error) {

	b := domain2.Budget{
		Category: req.Category,
		Limit:    decimal.NewFromFloat(req.Limit),
		Period:   req.Period,
	}

	if err := s.service.SetBudget(ctx, b); err != nil {
		return nil, mapDomainError(err)
	}

	return &ledgerv1.Budget{
		Category: b.Category,
		Limit:    req.Limit,
		Period:   b.Period,
	}, nil
}

func (s *Server) ListBudgets(
	ctx context.Context,
	_ *emptypb.Empty,
) (*ledgerv1.ListBudgetsResponse, error) {

	budgets, err := s.service.ListBudgets(ctx)
	if err != nil {
		return nil, mapDomainError(err)
	}

	res := &ledgerv1.ListBudgetsResponse{}
	for _, b := range budgets {
		res.Budgets = append(res.Budgets, &ledgerv1.Budget{
			Category: b.Category,
			Limit:    b.Limit.InexactFloat64(),
			Period:   b.Period,
		})
	}

	return res, nil
}

func (s *Server) GetReportSummary(
	ctx context.Context,
	req *ledgerv1.ReportSummaryRequest,
) (*ledgerv1.ReportSummaryResponse, error) {

	from, err := time.Parse("2006-01-02", req.From)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid from")
	}

	to, err := time.Parse("2006-01-02", req.To)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid to")
	}

	summary, err := s.service.GetReportSummary(ctx, from, to)
	if err != nil {
		return nil, mapDomainError(err)
	}

	resp := &ledgerv1.ReportSummaryResponse{
		Totals: make(map[string]float64),
	}

	for _, s := range summary {
		resp.Totals[s.Category] = s.Total.InexactFloat64()
	}

	return resp, nil
}

func (s *Server) BulkAddTransactions(
	ctx context.Context,
	req *ledgerv1.BulkAddTransactionsRequest,
) (*ledgerv1.BulkAddTransactionsResponse, error) {

	txs := make([]domain2.Transaction, 0, len(req.Transactions))

	for _, t := range req.Transactions {
		date, err := time.Parse("2006-01-02", t.Date)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid date")
		}

		txs = append(txs, domain2.Transaction{
			Amount:      decimal.NewFromFloat(t.Amount),
			Category:    t.Category,
			Description: t.Description,
			Date:        date,
		})
	}

	res, err := s.service.BulkAddTransactions(ctx, txs, int(req.Workers))
	if err != nil {
		return nil, mapDomainError(err)
	}

	out := &ledgerv1.BulkAddTransactionsResponse{
		Accepted: res.Accepted,
		Rejected: res.Rejected,
	}

	for _, e := range res.Errors {
		out.Errors = append(out.Errors, &ledgerv1.BulkError{
			Index: int32(e.Index),
			Error: e.Error,
		})
	}

	return out, nil
}
