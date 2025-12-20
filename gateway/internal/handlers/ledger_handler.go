package handlers

import (
	"encoding/csv"
	"encoding/json"
	"gateway/internal"
	"gateway/internal/middleware"
	ledgerv1 "gateway/ledger/v1"
	"net/http"
	"strconv"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Handler struct {
	client ledgerv1.LedgerServiceClient
}

func NewHandler(c ledgerv1.LedgerServiceClient) *Handler {
	return &Handler{client: c}
}

func responseJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}
func grpcErrorToHTTP(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		responseJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "internal_error",
		})
		return
	}

	switch st.Code() {
	case codes.InvalidArgument:
		responseJSON(w, http.StatusBadRequest, map[string]string{
			"error": st.Message(),
		})

	case codes.FailedPrecondition, codes.Aborted:
		responseJSON(w, http.StatusConflict, map[string]string{
			"error": st.Message(),
		})

	case codes.DeadlineExceeded:
		responseJSON(w, http.StatusGatewayTimeout, map[string]string{
			"error": "request timeout",
		})

	default:
		responseJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "internal_error",
		})
	}
}

// CreateTransaction godoc
// @Summary Create transaction
// @Tags transactions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body internal.CreateTransactionRequest true "Transaction"
// @Success 201 {object} map[string]bool
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/transactions [post]
func (h *Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		http.Error(
			w,
			"Content-Type must be application/json",
			http.StatusUnsupportedMediaType,
		)
		return
	}
	var dto internal.CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(
			w,
			"invalid json",
			http.StatusBadRequest,
		)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	md := metadata.New(map[string]string{
		"user_id": userID,
	})

	ctx := metadata.NewOutgoingContext(r.Context(), md)

	req := &ledgerv1.CreateTransactionRequest{
		Amount:      dto.Amount,
		Category:    dto.Category,
		Description: dto.Description,
		Date:        dto.Date,
	}

	_, err := h.client.AddTransaction(ctx, req)
	if err != nil {
		grpcErrorToHTTP(w, err)
		return
	}

	responseJSON(w, http.StatusCreated, map[string]bool{"success": true})
}

// ListTransactions godoc
// @Summary List transactions
// @Tags transactions
// @Security BearerAuth
// @Produce json
// @Success 200 {array} internal.TransactionResponse
// @Router /api/transactions [get]
func (h *Handler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	md := metadata.New(map[string]string{
		"user_id": userID,
	})

	ctx := metadata.NewOutgoingContext(r.Context(), md)

	resp, err := h.client.ListTransactions(
		ctx,
		&emptypb.Empty{},
	)
	if err != nil {
		grpcErrorToHTTP(w, err)
		return
	}

	out := make([]internal.TransactionResponse, 0, len(resp.Transactions))
	for _, t := range resp.Transactions {
		out = append(out, internal.TransactionResponse{
			Amount:      t.Amount,
			Category:    t.Category,
			Description: t.Description,
			Date:        t.Date,
		})
	}

	responseJSON(w, http.StatusOK, out)
}

// ListBudget godoc
// @Summary List budgets
// @Tags budgets
// @Security BearerAuth
// @Produce json
// @Success 200 {array} internal.BudgetResponse
// @Router /api/budgets [get]
func (h *Handler) ListBudget(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	md := metadata.New(map[string]string{
		"user_id": userID,
	})

	ctx := metadata.NewOutgoingContext(r.Context(), md)

	resp, err := h.client.ListBudgets(
		ctx,
		&emptypb.Empty{},
	)
	if err != nil {
		grpcErrorToHTTP(w, err)
		return
	}

	out := make([]internal.BudgetResponse, 0, len(resp.Budgets))
	for _, b := range resp.Budgets {
		out = append(out, internal.BudgetResponse{
			Category: b.Category,
			Limit:    b.Limit,
			Period:   b.Period,
		})
	}

	responseJSON(w, http.StatusOK, out)
}

// CreateBudget godoc
// @Summary Create budget
// @Tags budgets
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body internal.CreateBudgetRequest true "Budget"
// @Success 201 {object} map[string]bool
// @Router /api/budgets [post]
func (h *Handler) CreateBudget(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		http.Error(
			w,
			"Content-Type must be application/json",
			http.StatusUnsupportedMediaType,
		)
		return
	}
	var dto internal.CreateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		responseJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid json",
		})
		return
	}

	req := &ledgerv1.CreateBudgetRequest{
		Category: dto.Category,
		Limit:    dto.Limit,
		Period:   dto.Period,
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	md := metadata.New(map[string]string{
		"user_id": userID,
	})

	ctx := metadata.NewOutgoingContext(r.Context(), md)

	_, err := h.client.SetBudget(ctx, req)
	if err != nil {
		grpcErrorToHTTP(w, err)
		return
	}

	responseJSON(w, http.StatusCreated, map[string]bool{"success": true})
}

// ReportSummary godoc
// @Summary Expense summary
// @Tags reports
// @Security BearerAuth
// @Produce json
// @Param from query string true "From date (YYYY-MM-DD)"
// @Param to query string true "To date (YYYY-MM-DD)"
// @Success 200 {object} map[string]float64
// @Router /api/reports/summary [get]
func (h *Handler) ReportSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	req := &ledgerv1.ReportSummaryRequest{
		From: r.URL.Query().Get("from"),
		To:   r.URL.Query().Get("to"),
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	md := metadata.New(map[string]string{
		"user_id": userID,
	})

	ctx := metadata.NewOutgoingContext(r.Context(), md)

	resp, err := h.client.GetReportSummary(ctx, req)
	if err != nil {
		grpcErrorToHTTP(w, err)
		return
	}

	responseJSON(w, http.StatusOK, resp.Totals)
}

func respondTimeout(w http.ResponseWriter) {
	responseJSON(w, http.StatusGatewayTimeout, map[string]string{
		"error": "request timeout",
	})
}

type BulkAddTransactionsResponse struct {
	Success int64 `json:"success"`
	Failed  int64 `json:"failed"`
}

// BulkCreateTransactions godoc
// @Summary Bulk create transactions
// @Tags transactions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body []internal.CreateTransactionRequest true "Transactions"
// @Success 200 {object} BulkAddTransactionsResponse
// @Router /api/transactions/bulk [post]
func (h *Handler) BulkCreateTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		http.Error(
			w,
			"Content-Type must be application/json",
			http.StatusUnsupportedMediaType,
		)
		return
	}
	var dtos []internal.CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&dtos); err != nil {
		responseJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid json",
		})
		return
	}

	workers := int32(4)
	if v := r.URL.Query().Get("workers"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			workers = int32(n)
		}
	}

	req := &ledgerv1.BulkAddTransactionsRequest{
		Workers: workers,
	}

	for _, d := range dtos {
		req.Transactions = append(req.Transactions, &ledgerv1.CreateTransactionRequest{
			Amount:      d.Amount,
			Category:    d.Category,
			Description: d.Description,
			Date:        d.Date,
		})
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	md := metadata.New(map[string]string{
		"user_id": userID,
	})

	ctx := metadata.NewOutgoingContext(r.Context(), md)

	resp, err := h.client.BulkAddTransactions(ctx, req)
	if err != nil {
		grpcErrorToHTTP(w, err)
		return
	}

	responseJSON(w, http.StatusOK, BulkAddTransactionsResponse{
		Success: resp.Accepted,
		Failed:  resp.Rejected,
	})
}

func (h *Handler) ImportCSV(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	md := metadata.New(map[string]string{
		"user_id": userID,
	})

	ctx := metadata.NewOutgoingContext(r.Context(), md)

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		http.Error(w, "cannot parse form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	rows, err := reader.ReadAll()
	if err != nil {
		http.Error(w, "invalid csv", http.StatusBadRequest)
		return
	}

	if len(rows) < 2 {
		http.Error(w, "empty csv", http.StatusBadRequest)
		return
	}

	var txs []*ledgerv1.CreateTransactionRequest
	for i, row := range rows[1:] {
		if len(row) < 4 {
			continue
		}

		amount, err := strconv.ParseFloat(row[0], 64)
		if err != nil {
			continue
		}

		txs = append(txs, &ledgerv1.CreateTransactionRequest{
			Amount:      amount,
			Category:    row[1],
			Description: row[2],
			Date:        row[3],
		})

		_ = i
	}

	resp, err := h.client.BulkAddTransactions(
		ctx,
		&ledgerv1.BulkAddTransactionsRequest{
			Transactions: txs,
			Workers:      4,
		},
	)
	if err != nil {
		grpcErrorToHTTP(w, err)
		return
	}
	responseJSON(w, http.StatusOK, resp)
}

func (h *Handler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	md := metadata.New(map[string]string{
		"user_id": userID,
	})

	ctx := metadata.NewOutgoingContext(r.Context(), md)

	resp, err := h.client.ListTransactions(
		ctx,
		&emptypb.Empty{},
	)
	if err != nil {
		grpcErrorToHTTP(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set(
		"Content-Disposition",
		`attachment; filename="transactions.csv"`,
	)

	writer := csv.NewWriter(w)
	defer writer.Flush()

	_ = writer.Write([]string{
		"amount", "category", "description", "date",
	})

	for _, t := range resp.Transactions {
		_ = writer.Write([]string{
			strconv.FormatFloat(t.Amount, 'f', 2, 64),
			t.Category,
			t.Description,
			t.Date,
		})
	}
}
